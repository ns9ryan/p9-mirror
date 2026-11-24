package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type ClaimAgentDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClaimAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimAgentDepositOrderLogic {
	return &ClaimAgentDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 领取审核
func (l *ClaimAgentDepositOrderLogic) ClaimAgentDepositOrder(in *agentdepositorderpb.ClaimAgentDepositOrderReq) (*agentdepositorderpb.ClaimAgentDepositOrderResp, error) {
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	resp, err := l.svcCtx.Agent.ClaimAgentDepositOrder(l.ctx, in.OrderId)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
