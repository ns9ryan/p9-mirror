package memberlogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	agentdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/agent"
	memberdao "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

const registerChannelH5 = "h5"

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 使用 H5 API 解析的厅及代理归属注册会员。
func (l *RegisterLogic) Register(in *memberpb.RegisterMemberRequest) (*memberpb.RegisterMemberResponse, error) {
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if strings.TrimSpace(operatorCode) == "" {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	account := strings.TrimSpace(in.GetAccount())
	password := in.GetPassword()
	agentID := in.GetAgentId()
	if !validMemberAccount(account) || !validMemberPassword(password) || agentID < 0 {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	passwordHash, err := dao.HashPassword(password)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	var created *ent.Member
	err = dao.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		// 未指定代理时，读取当前厅代理配置里的现金生产代理。
		useDefaultAgent := agentID == 0
		if useDefaultAgent {
			defaultAgentID, err := agentdao.DefaultMemberRegistrationAgentIDTx(l.ctx, tx)
			if err != nil {
				return fmt.Errorf("查询默认注册代理失败: %w", err)
			}
			if defaultAgentID <= 0 {
				return status.Error(codes.Unavailable, "default registration agent is not configured")
			}
			agentID = defaultAgentID
		}

		// 校验代理属于当前厅且可接纳会员，默认代理须为二级。
		agentQuery := tx.AgentUser.Query().Where(
			agentuser.IDEQ(agentID),
			agentuser.OperatorCodeEQ(operatorCode),
			agentuser.LevelGTE(1),
			agentuser.StatusEQ(memberAgentStatusNormal),
		)
		if useDefaultAgent {
			agentQuery.Where(agentuser.LevelEQ(2))
		}
		validAgent, err := agentQuery.Exist(l.ctx)
		if err != nil {
			return fmt.Errorf("校验注册代理失败: %w", err)
		}
		if !validAgent {
			if useDefaultAgent {
				return status.Error(codes.Unavailable, "default registration agent is invalid")
			}
			return xerr.BadRequest(i18n.InvalidParam)
		}

		exists, err := l.svcCtx.Member.UsernameExistsTx(l.ctx, tx, account)
		if err != nil {
			return fmt.Errorf("检查会员账号是否重复失败: %w", err)
		}
		if exists {
			return xerr.BadRequest(i18n.InvalidParam)
		}

		code, err := generateMemberCode(l.ctx, tx, l.svcCtx.Member)
		if err != nil {
			return err
		}
		inviteCode, err := generateMemberInviteCode(l.ctx, tx, l.svcCtx.Member)
		if err != nil {
			return err
		}

		created, err = l.svcCtx.Member.CreateIdentityTx(l.ctx, tx, memberdao.CreateParams{
			Code:            code,
			InviteCode:      inviteCode,
			AgentID:         agentID,
			Account:         &account,
			PasswordHash:    &passwordHash,
			RegisterMethod:  registerMethodAccount,
			RegisterChannel: registerChannelH5,
		})
		if err != nil {
			return err
		}

		// 会员资料、会员统计及 VIP 晋级进度待对应业务接入后启用。
		// if err := l.svcCtx.Member.CreateProfileTx(l.ctx, tx, memberdao.CreateParams{
		//     Nickname:      account,
		//     AvatarID:      defaultMemberAvatarID,
		//     AvatarFrameID: defaultMemberAvatarFrameID,
		// }, created.ID); err != nil {
		//     return err
		// }
		// if err := l.svcCtx.Member.CreateStatsTx(l.ctx, tx, created.ID, created.CreatedAt); err != nil {
		//     return err
		// }
		// if err := l.svcCtx.Member.CreateVipProgressTx(l.ctx, tx, created.ID); err != nil {
		//     return err
		// }

		if err := agentdao.BumpMemberCounts(l.ctx, tx, agentID); err != nil {
			return fmt.Errorf("更新代理会员统计失败: %w", err)
		}
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

	l.svcCtx.KafkaEventHandler.PublishMemberRegisteredEvent(l.ctx, created)
	return &memberpb.RegisterMemberResponse{Id: created.ID, Code: created.Code}, nil
}
