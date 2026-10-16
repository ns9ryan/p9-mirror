package agentservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TreeNodeAgentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTreeNodeAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreeNodeAgentsLogic {
	return &TreeNodeAgentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: agent
func (l *TreeNodeAgentsLogic) TreeNodeAgents(in *agentpb.TreeNodeAgentsReq) (*agentpb.TreeNodeAgentsResp, error) {
	resp, err := l.svcCtx.Agent.TreeNodeAgents(l.ctx, in)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	return resp, nil
}
