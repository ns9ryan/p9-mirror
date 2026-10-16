package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAgentAppliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAgentAppliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentAppliesLogic {
	return &ListAgentAppliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAgentApplies 申请代理列表
func (l *ListAgentAppliesLogic) ListAgentApplies(in *agentpb.ListAgentAppliesReq) (*agentpb.ListAgentAppliesResp, error) {
	resp, err := l.svcCtx.Agent.ListAgentApplies(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
