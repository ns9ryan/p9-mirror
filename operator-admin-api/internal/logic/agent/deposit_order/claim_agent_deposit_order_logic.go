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

type ClaimAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 领取审核
func NewClaimAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimAgentDepositOrderLogic {
	return &ClaimAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClaimAgentDepositOrderLogic) ClaimAgentDepositOrder(req *types.AgentDepositOrderActionReq) (*types.AgentDepositOrderActionResp, error) {
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.ClaimAgentDepositOrder(l.ctx, &agentdepositorderpb.ClaimAgentDepositOrderReq{
		OrderId: req.OrderId,
	})
	if err != nil {
		return nil, err
	}
	return &types.AgentDepositOrderActionResp{
		OrderId: rpcResp.GetOrderId(),
		Status:  rpcResp.GetStatus(),
	}, nil
}
