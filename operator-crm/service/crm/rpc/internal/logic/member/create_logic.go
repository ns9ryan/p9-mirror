package memberlogic

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	agentdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/agent"
	memberdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

const (
	memberCodeMin            int64 = 100_000_000_000 // 12位会员编号的最小值
	memberCodeRange          int64 = 900_000_000_000 // 12位会员编号的取值范围
	memberCodeAttempts             = 10              // 会员编号生成重试次数
	memberInviteCodeAttempts       = 10              // 会员邀请码生成重试次数
	memberAgentStatusNormal  int64 = 1               // 正常代理状态

	defaultMemberAvatarID      int64 = 0 // 0表示尚未选择总站头像资源
	defaultMemberAvatarFrameID int64 = 0 // 0表示尚未选择总站头像框资源

	registerMethodAccount = "account" // TODO: 注册方式枚举统一后替换临时值
	registerChannelAdmin  = "admin"   // TODO: 注册渠道枚举统一后替换临时值
)

// 会员账号仅允许英文字母, 数字, 横杠和下横杠
var memberAccountPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// 会员密码仅允许ASCII可见字符, 不包含空格
var memberPasswordPattern = regexp.MustCompile(`^[\x21-\x7E]+$`)

// CreateLogic 会员创建逻辑
type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateLogic 创建会员逻辑
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create 创建会员
func (l *CreateLogic) Create(in *memberpb.CreateMemberRequest) (*memberpb.CreateMemberResponse, error) {
	// 校验分站运营身份和当前厅
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 规范化并校验输入
	account := strings.TrimSpace(in.GetAccount())
	password := in.GetPassword()
	if !validMemberAccount(account) || !validMemberPassword(password) || in.GetAgentId() <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 处理会员昵称
	nickname, err := l.resolveNickname(in.GetNickname(), account)
	if err != nil {
		return nil, err
	}

	// 处理后台会员备注
	var remark *string
	if in.Remark != nil {
		value := strings.TrimSpace(in.GetRemark())
		if utf8.RuneCountInString(value) > 200 {
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
		}
		if value != "" {
			remark = &value
		}
	}

	// 校验所属代理的状态和层级
	agent, err := l.svcCtx.Agent.LoadAgent(l.ctx, in.GetAgentId())
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	if agent.Level < 1 || agent.Status != memberAgentStatusNormal {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 生成登录密码哈希
	passwordHash, err := dao.HashPassword(password)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	// 会员及代理统计必须在同一事务中更新
	var created *ent.Member
	err = dao.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		// 检查当前厅的登录账号
		exists, err := l.svcCtx.Member.UsernameExistsTx(l.ctx, tx, account)
		if err != nil {
			return fmt.Errorf("检查会员账号是否重复失败: %w", err)
		}
		if exists {
			return xerr.BadRequest(i18n.InvalidParam)
		}

		createdMember, err := l.createMemberTx(tx, memberdao.CreateParams{
			AgentID:         in.GetAgentId(),            // 会员所属分层代理ID
			Account:         &account,                   // 会员登录账号
			PasswordHash:    &passwordHash,              // 登录密码单向哈希
			Nickname:        nickname,                   // 会员昵称
			AvatarID:        defaultMemberAvatarID,      // 会员头像URL
			AvatarFrameID:   defaultMemberAvatarFrameID, // 会员头像框URL
			Remark:          remark,                     // 后台会员备注
			RegisterMethod:  registerMethodAccount,      // 注册方式
			RegisterChannel: registerChannelAdmin,       // 注册渠道
		})
		if err != nil {
			return err
		}

		created = createdMember
		return nil
	})
	if err != nil {
		switch {
		case ent.IsConstraintError(err):
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		case ent.IsValidationError(err):
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		default:
			return nil, xerr.RpcErr(err)
		}
	}

	// 发送会员注册事件到 Kafka
	l.svcCtx.KafkaEventHandler.PublishMemberRegisteredEvent(l.ctx, created)

	// 返回会员ID和对外编号
	return &memberpb.CreateMemberResponse{
		Id:   created.ID,
		Code: created.Code,
	}, nil
}

// validMemberAccount 校验会员账号格式
func validMemberAccount(account string) bool {
	return len(account) >= 4 && len(account) <= 20 && memberAccountPattern.MatchString(account)
}

// validMemberPassword 校验会员密码格式
func validMemberPassword(password string) bool {
	return len(password) >= 8 && len(password) <= 32 && memberPasswordPattern.MatchString(password)
}

// generateMemberCode 生成当前厅唯一的12位会员编号
func generateMemberCode(ctx context.Context, tx *ent.Tx, memberDAO *memberdao.Service) (string, error) {
	for i := 0; i < memberCodeAttempts; i++ {
		code := strconv.FormatInt(memberCodeMin+rand.Int64N(memberCodeRange), 10)

		exists, err := memberDAO.CodeExistsTx(ctx, tx, code)
		if err != nil {
			return "", fmt.Errorf("检查会员编号是否重复失败: %w", err)
		}
		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("生成会员编号失败: 已达到最大重试次数")
}

// generateMemberInviteCode 生成当前厅唯一的会员邀请码
func generateMemberInviteCode(ctx context.Context, tx *ent.Tx, memberDAO *memberdao.Service) (string, error) {
	for i := 0; i < memberInviteCodeAttempts; i++ {
		code, err := utils.GenShortID()
		if err != nil {
			return "", fmt.Errorf("生成会员邀请码失败: %w", err)
		}
		code = strings.ToUpper(code)

		exists, err := memberDAO.InviteCodeExistsTx(ctx, tx, code)
		if err != nil {
			return "", fmt.Errorf("检查会员邀请码是否重复失败: %w", err)
		}
		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("生成会员邀请码失败: 已达到最大重试次数")
}

// 显式传入的昵称会保留，空值才查询昵称池
func (l *CreateLogic) resolveNickname(value, account string) (string, error) {
	nickname := strings.TrimSpace(value)
	if nickname == "" {
		// 昵称库为空时沿用账号兜底，RPC 异常明确返回。
		if l.svcCtx.Nicknames == nil {
			return "", status.Error(codes.Unavailable, "platform-base nickname RPC is not configured")
		}
		result, err := l.svcCtx.Nicknames.Random(l.ctx, &nicknamepb.RandomNicknameRequest{})
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return "", err
			}
			nickname = account
		} else {
			nickname = strings.TrimSpace(result.GetNickname().GetName())
			if nickname == "" {
				return "", status.Error(codes.Internal, "platform-base returned an empty nickname")
			}
		}
	}
	if utf8.RuneCountInString(nickname) > 100 {
		return "", xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	return nickname, nil
}

func (l *CreateLogic) createMemberTx(tx *ent.Tx, params memberdao.CreateParams) (*ent.Member, error) {
	// 生成唯一会员编号
	code, err := generateMemberCode(l.ctx, tx, l.svcCtx.Member)
	if err != nil {
		return nil, err
	}

	// 生成唯一会员邀请码
	inviteCode, err := generateMemberInviteCode(l.ctx, tx, l.svcCtx.Member)
	if err != nil {
		return nil, err
	}

	params.Code = code
	params.InviteCode = inviteCode
	createdMember, err := l.svcCtx.Member.CreateTx(l.ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("创建会员失败: %w", err)
	}

	// 更新所属代理及其上级的会员统计
	if err := agentdao.BumpMemberCounts(l.ctx, tx, params.AgentID); err != nil {
		return nil, fmt.Errorf("更新代理会员统计失败: %w", err)
	}

	return createdMember, nil
}
