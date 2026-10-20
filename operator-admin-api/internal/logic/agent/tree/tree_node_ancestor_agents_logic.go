package tree

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TreeNodeAncestorAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTreeNodeAncestorAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreeNodeAncestorAgentsLogic {
	return &TreeNodeAncestorAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TreeNodeAncestorAgentsLogic) TreeNodeAncestorAgents(req *types.TreeNodeAncestorAgentsReq) ([]types.AgentSearchItem, error) {
	if req == nil {
		req = &types.TreeNodeAncestorAgentsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.TreeNodeAncestorAgents(l.ctx, &agentpb.TreeNodeAncestorAgentsReq{
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return agentconv.SearchItemsFromPB(rpcResp.GetItems()), nil
}
