package agentdepositorderservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentDepositOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAgentDepositOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentDepositOrdersLogic {
	return &ListAgentDepositOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 存款订单列表
func (l *ListAgentDepositOrdersLogic) ListAgentDepositOrders(in *agentdepositorderpb.ListAgentDepositOrdersReq) (*agentdepositorderpb.ListAgentDepositOrdersResp, error) {
	resp, err := l.svcCtx.Agent.ListAgentDepositOrders(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
