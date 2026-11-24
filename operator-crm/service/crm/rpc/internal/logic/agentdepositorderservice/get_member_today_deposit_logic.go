package agentdepositorderservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemberTodayDepositLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemberTodayDepositLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberTodayDepositLogic {
	return &GetMemberTodayDepositLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 获取会员今日累计有效充值金额
func (l *GetMemberTodayDepositLogic) GetMemberTodayDeposit(in *agentdepositorderpb.GetMemberTodayDepositReq) (*agentdepositorderpb.GetMemberTodayDepositResp, error) {
	if in == nil || in.MemberId <= 0 {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}

	amount, count, err := l.svcCtx.Agent.GetMemberTodayDeposit(l.ctx, in.MemberId)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}

	return &agentdepositorderpb.GetMemberTodayDepositResp{
		TodayDepositAmount: amount,
		TodayDepositCount:  count,
	}, nil
}

