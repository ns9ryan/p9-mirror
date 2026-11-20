package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOperatorAgentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOperatorAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOperatorAgentsLogic {
	return &ListOperatorAgentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOperatorAgents 按层级列出代理并按线路分组
func (l *ListOperatorAgentsLogic) ListOperatorAgents(in *agentpb.ListOperatorAgentsReq) (*agentpb.ListOperatorAgentsResp, error) {
	resp, err := l.svcCtx.Agent.ListOperatorAgents(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
