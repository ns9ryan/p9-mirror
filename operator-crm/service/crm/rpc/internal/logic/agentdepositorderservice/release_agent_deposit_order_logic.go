package agentdepositorderservicelogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentdepositorderpb"
)

type ReleaseAgentDepositOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseAgentDepositOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseAgentDepositOrderLogic {
	return &ReleaseAgentDepositOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent_deposit_order 释放审核
func (l *ReleaseAgentDepositOrderLogic) ReleaseAgentDepositOrder(in *agentdepositorderpb.ReleaseAgentDepositOrderReq) (*agentdepositorderpb.ReleaseAgentDepositOrderResp, error) {
	if in == nil || in.OrderId <= 0 {
		return nil, xerr.RpcErr(xerr.BadRequest("common.invalidParam"))
	}
	resp, err := l.svcCtx.Agent.ReleaseAgentDepositOrder(l.ctx, in.OrderId)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
