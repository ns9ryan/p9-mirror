package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type UnfreezeAgentDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfreezeAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeAgentDepositOrderLogic {
	return &UnfreezeAgentDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 解冻订单
func (l *UnfreezeAgentDepositOrderLogic) UnfreezeAgentDepositOrder(in *agentdepositorderpb.UnfreezeAgentDepositOrderReq) (*agentdepositorderpb.UnfreezeAgentDepositOrderResp, error) {
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	resp, err := l.svcCtx.Agent.UnfreezeAgentDepositOrder(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
