package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TreeNodeAncestorAgentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTreeNodeAncestorAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreeNodeAncestorAgentsLogic {
	return &TreeNodeAncestorAgentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TreeNodeAncestorAgentsLogic) TreeNodeAncestorAgents(in *agentpb.TreeNodeAncestorAgentsReq) (*agentpb.TreeNodeAncestorAgentsResp, error) {
	resp, err := l.svcCtx.Agent.TreeNodeAncestorAgents(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
