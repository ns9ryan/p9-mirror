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

// ListOperatorAgents 按线路分组获取指定层级的代理商列表
func (l *ListOperatorAgentsLogic) ListOperatorAgents(req *types.ListOperatorAgentsReq) (*types.ListOperatorAgentsResp, error) {
	if req == nil {
		req = &types.ListOperatorAgentsReq{}
	}
	// 获取已开通的线路
	openedLines, err := agentconv.OpenedLineCodes(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	// 获取线路下的代理商
	rpcResp, err := l.svcCtx.AgentRpc.ListOperatorAgents(l.ctx, &agentpb.ListOperatorAgentsReq{
		Level: req.Level,
	})
	if err != nil {
		return nil, err
	}
	// 按线路分组
	byLine := make(map[string][]types.AgentSearchItem, len(rpcResp.GetGroups()))
	for _, g := range rpcResp.GetGroups() {
		if g == nil {
			continue
		}
		byLine[g.LineCode] = agentconv.SearchItemsFromPB(g.Agents)
	}
	groups := make([]types.OperatorAgentLineGroup, 0, len(openedLines))
	for _, code := range openedLines {
		agents := byLine[code]
		if agents == nil {
			agents = []types.AgentSearchItem{}
		}
		groups = append(groups, types.OperatorAgentLineGroup{
			LineCode: code,
			Agents:   agents,
		})
	}
	return &types.ListOperatorAgentsResp{Groups: groups}, nil
}
