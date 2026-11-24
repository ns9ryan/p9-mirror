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

type RejectAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审核拒绝
func NewRejectAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectAgentDepositOrderLogic {
	return &RejectAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RejectAgentDepositOrderLogic) RejectAgentDepositOrder(req *types.AgentDepositOrderRejectReq) (*types.AgentDepositOrderActionResp, error) {
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.RejectAgentDepositOrder(l.ctx, &agentdepositorderpb.RejectAgentDepositOrderReq{
		OrderId:    req.OrderId,
		AuditNotes: req.Notes,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentDepositOrderActionResp{
		OrderId: rpcResp.GetOrderId(),
		Status:  rpcResp.GetStatus(),
	}, nil
}
