// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tree

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TreeNodeAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 代理树节点（按上级 ID；0 为顶层）
func NewTreeNodeAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreeNodeAgentsLogic {
	return &TreeNodeAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TreeNodeAgentsLogic) TreeNodeAgents(req *types.TreeNodeAgentsReq) ([]types.TreeNodeAgentItem, error) {
	if req == nil {
		req = &types.TreeNodeAgentsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.TreeNodeAgents(l.ctx, &agentpb.TreeNodeAgentsReq{
		ParentId: req.ParentId,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.TreeNodeAgentItem, 0, len(rpcResp.GetItems()))
	for _, it := range rpcResp.GetItems() {
		if it == nil {
			continue
		}
		items = append(items, types.TreeNodeAgentItem{
			Id:               it.Id,
			Account:          it.Account,
			LineCode:         it.LineCode,
			Level:            it.Level,
			LevelAlias:       it.LevelAlias,
			DirectAgentCount: it.DirectAgentCount,
		})
	}
	return items, nil
}
