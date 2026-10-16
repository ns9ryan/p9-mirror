package manage

import (
	"context"

	agentconv "oa.98ent.com/p9/operator-admin-api/internal/logic/agent"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOperatorAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOperatorAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOperatorAgentsLogic {
	return &ListOperatorAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOperatorAgentsLogic) ListOperatorAgents(req *types.ListOperatorAgentsReq) (*types.ListOperatorAgentsResp, error) {
	if req == nil {
		req = &types.ListOperatorAgentsReq{}
	}
	rpcResp, err := l.svcCtx.AgentRpc.ListOperatorAgents(l.ctx, &agentpb.ListOperatorAgentsReq{
		Level: req.Level,
	})
	if err != nil {
		return nil, err
	}
	groups := make([]types.OperatorAgentLineGroup, 0, len(rpcResp.GetGroups()))
	for _, g := range rpcResp.GetGroups() {
		if g == nil {
			continue
		}
		groups = append(groups, types.OperatorAgentLineGroup{
			LineCode: g.LineCode,
			Agents:   agentconv.SearchItemsFromPB(g.Agents),
		})
	}
	return &types.ListOperatorAgentsResp{Groups: groups}, nil
}
