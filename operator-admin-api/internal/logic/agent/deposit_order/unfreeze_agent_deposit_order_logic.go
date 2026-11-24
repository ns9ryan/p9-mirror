// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package deposit_order

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfreezeAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解冻订单
func NewUnfreezeAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeAgentDepositOrderLogic {
	return &UnfreezeAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnfreezeAgentDepositOrderLogic) UnfreezeAgentDepositOrder(req *types.AgentDepositOrderAuditReq) (*types.AgentDepositOrderActionResp, error) {
	pbReq := &agentdepositorderpb.UnfreezeAgentDepositOrderReq{
		OrderId: req.OrderId,
	}
	if req.Notes != "" {
		pbReq.UnfreezeNotes = &req.Notes
	}
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.UnfreezeAgentDepositOrder(l.ctx, pbReq)
	if err != nil {
		return nil, err
	}
	return &types.AgentDepositOrderActionResp{
		OrderId: rpcResp.GetOrderId(),
		Status:  rpcResp.GetStatus(),
	}, nil
}
