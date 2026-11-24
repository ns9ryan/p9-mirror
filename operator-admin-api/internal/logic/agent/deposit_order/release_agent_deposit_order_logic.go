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

type ReleaseAgentDepositOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 释放审核
func NewReleaseAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseAgentDepositOrderLogic {
	return &ReleaseAgentDepositOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseAgentDepositOrderLogic) ReleaseAgentDepositOrder(req *types.AgentDepositOrderActionReq) (*types.AgentDepositOrderActionResp, error) {
	rpcResp, err := l.svcCtx.AgentDepositOrderRpc.ReleaseAgentDepositOrder(l.ctx, &agentdepositorderpb.ReleaseAgentDepositOrderReq{
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
