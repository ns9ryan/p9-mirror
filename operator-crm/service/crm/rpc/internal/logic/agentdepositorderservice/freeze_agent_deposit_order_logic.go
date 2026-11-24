package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type FreezeAgentDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFreezeAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeAgentDepositOrderLogic {
	return &FreezeAgentDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 冻结订单
func (l *FreezeAgentDepositOrderLogic) FreezeAgentDepositOrder(in *agentdepositorderpb.FreezeAgentDepositOrderReq) (*agentdepositorderpb.FreezeAgentDepositOrderResp, error) {
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	resp, err := l.svcCtx.Agent.FreezeAgentDepositOrder(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
