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

type FreezeAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 冻结订单
func NewFreezeAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeAgentDepositOrderLogic {
	return &FreezeAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FreezeAgentDepositOrderLogic) FreezeAgentDepositOrder(req *types.AgentDepositOrderAuditReq) (*types.AgentDepositOrderActionResp, error) {
	pbReq := &agentdepositorderpb.FreezeAgentDepositOrderReq{
		OrderId: req.OrderId,
	}
	if req.Notes != "" {
		pbReq.FreezeNotes = &req.Notes
	}
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.FreezeAgentDepositOrder(l.ctx, pbReq)
	if err != nil {
		return nil, err
	}
	return &types.AgentDepositOrderActionResp{
		OrderId: rpcResp.GetOrderId(),
		Status:  rpcResp.GetStatus(),
	}, nil
}
