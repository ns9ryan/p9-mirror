package memberlogic

import (
	"context"
	"math/rand/v2"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	batchMemberLimit           = 2000
	batchMemberPassword        = "12345678"
	batchMemberAccountAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	batchMemberAccountAttempts = 32
)

type BatchCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCreateLogic {
	return &BatchCreateLogic{ctx: ctx, svcCtx: svcCtx}
}

// BatchCreate 批量创建当前分站会员。
func (l *BatchCreateLogic) BatchCreate(in *memberpb.BatchCreateMembersRequest) (*memberpb.BatchCreateMembersResponse, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 校验前后缀和创建数量，并为随机字符预留至少 4 位。
	prefix, suffix := in.GetAccountPrefix(), in.GetAccountSuffix()
	count := in.GetCount()
	if count < 1 || count > batchMemberLimit || in.GetAgentId() <= 0 ||
		len(prefix)+len(suffix) == 0 || len(prefix)+len(suffix) > 16 ||
		(prefix != "" && !memberAccountPattern.MatchString(prefix)) ||
		(suffix != "" && !memberAccountPattern.MatchString(suffix)) {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 批量创建只接受当前分站下状态正常的分层代理。
	agent, err := l.svcCtx.Agent.LoadAgent(l.ctx, in.GetAgentId())
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	if agent.Level < 1 || agent.Status != memberAgentStatusNormal {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	// 在数据库事务外准备昵称和密码哈希，减少事务占用时间。
	params, err := l.prepareMembers(count, in.GetAgentId())
	if err != nil {
		return nil, err
	}

	var members []*ent.Member
	// 所有分批写入与代理统计共用一个事务，任一环节失败则整单回滚。
	err = dao.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		if err := l.assignMemberIdentifiers(tx, params, prefix, suffix); err != nil {
			return err
		}
		created, err := l.svcCtx.Member.BatchCreateTx(l.ctx, tx, params)
		if err != nil {
			return err
		}
		if err := agentdao.AddMemberCounts(l.ctx, tx, in.GetAgentId(), int64(count)); err != nil {
			return err
		}
		members = created
		return nil
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, xerr.RpcErr(xerr.EntConstraintError(i18n.InvalidParam, err))
		}
		if ent.IsValidationError(err) {
			return nil, xerr.RpcErr(xerr.EntValidationError(i18n.InvalidParam, err))
		}
		return nil, xerr.RpcErr(err)
	}
	result := make([]*memberpb.BatchCreatedMember, len(members))
	for i, member := range members {
		result[i] = &memberpb.BatchCreatedMember{Id: member.ID, Code: member.Code, Account: *params[i].Account}
	}
	return &memberpb.BatchCreateMembersResponse{Members: result}, nil
}

func (l *BatchCreateLogic) prepareMembers(count int32, agentID int64) ([]memberdao.CreateParams, error) {
	params := make([]memberdao.CreateParams, count)
	group, ctx := errgroup.WithContext(l.ctx)
	// 最多四个并发任务，每个任务负责不同的下标，避免同时写同一条记录。
	workers := min(4, len(params))
	for worker := range workers {
		group.Go(func() error {
			create := NewCreateLogic(ctx, l.svcCtx)
			for i := worker; i < len(params); i += workers {
				if err := ctx.Err(); err != nil {
					return status.FromContextError(err).Err()
				}
				nickname, err := create.resolveNickname("", "")
				if err != nil {
					return err
				}
				if nickname == "" {
					return status.Error(codes.FailedPrecondition, "platform-base nickname pool is empty")
				}
				// 每个会员单独计算密码哈希，使用独立盐值。
				hash, err := dao.HashPassword(batchMemberPassword)
				if err != nil {
					return xerr.RpcErr(err)
				}
				params[i] = memberdao.CreateParams{
					AgentID: agentID, Nickname: nickname, PasswordHash: &hash,
					AvatarID: defaultMemberAvatarID, AvatarFrameID: defaultMemberAvatarFrameID,
					RegisterMethod: registerMethodAccount, RegisterChannel: registerChannelAdmin,
				}
			}
			return nil
		})
	}
	return params, group.Wait()
}

func (l *BatchCreateLogic) assignMemberIdentifiers(tx *ent.Tx, params []memberdao.CreateParams, prefix, suffix string) error {
	pending := make([]int, len(params))
	for i := range params {
		pending[i] = i
	}
	// 先在内存中排除本次请求的重复标识，再批量查询数据库。
	usedAccounts := make(map[string]bool, len(params))
	usedCodes := make(map[string]bool, len(params))
	usedInvites := make(map[string]bool, len(params))
	middle := make([]byte, min(8, 20-len(prefix)-len(suffix)))
	for attempt := 0; attempt < batchMemberAccountAttempts && len(pending) > 0; attempt++ {
		candidates := make([]int, 0, len(pending))
		retry := make([]int, 0)
		accounts := make([]string, 0, len(pending))
		codes := make([]string, 0, len(pending))
		invites := make([]string, 0, len(pending))
		for _, i := range pending {
			for j := range middle {
				middle[j] = batchMemberAccountAlphabet[rand.IntN(len(batchMemberAccountAlphabet))]
			}
			account := prefix + string(middle) + suffix
			code := strconv.FormatInt(memberCodeMin+rand.Int64N(memberCodeRange), 10)
			invite, err := utils.GenShortID()
			if err != nil {
				return err
			}
			invite = strings.ToUpper(invite)
			if usedAccounts[account] || usedCodes[code] || usedInvites[invite] {
				retry = append(retry, i)
				continue
			}
			usedAccounts[account], usedCodes[code], usedInvites[invite] = true, true, true
			params[i].Account, params[i].Code, params[i].InviteCode = &account, code, invite
			candidates = append(candidates, i)
			accounts, codes, invites = append(accounts, account), append(codes, code), append(invites, invite)
		}

		if len(candidates) > 0 {
			conflicts, err := l.svcCtx.Member.FindIdentifierConflictsTx(l.ctx, tx, accounts, codes, invites)
			if err != nil {
				return err
			}
			for j, i := range candidates {
				if conflicts.Accounts[strings.ToLower(accounts[j])] || conflicts.Codes[codes[j]] || conflicts.InviteCodes[invites[j]] {
					retry = append(retry, i)
				}
			}
		}
		// 已通过检查的记录保留原结果，只为冲突的记录重新生成。
		pending = retry
	}
	if len(pending) > 0 {
		return xerr.BadRequest(i18n.InvalidParam)
	}
	return nil
}
