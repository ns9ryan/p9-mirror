package agent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

// TreeNodeAgents 按上级 ID 查直属下级树；parent_id<=0 时返回顶层代理（level=0）。
func (s *Service) TreeNodeAgents(ctx context.Context, in *agentpb.TreeNodeAgentsReq) (*agentpb.TreeNodeAgentsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &agentpb.TreeNodeAgentsReq{}
	}

	q := s.DB.AgentUser.Query()
	if in.ParentId > 0 {
		q = q.Where(agentuser.ParentIDEQ(in.ParentId))
	} else {
		q = q.Where(agentuser.ParentIDIsNil())
	}

	rows, err := q.
		WithStats().
		Order(agentuser.ByID(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, err
	}

	levelAlias := make(map[int64]string)
	levels, err := s.DB.AgentLevel.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	for _, lvl := range levels {
		levelAlias[lvl.Level] = lvl.Alias
	}

	items := make([]*agentpb.TreeNodeAgentItem, 0, len(rows))
	for _, row := range rows {
		item := &agentpb.TreeNodeAgentItem{
			Id:         row.ID,
			Account:    row.Account,
			LineCode:   row.AgentLineCode,
			Level:      int32(row.Level),
			LevelAlias: levelAlias[row.Level],
		}
		if st := row.Edges.Stats; st != nil {
			item.DirectAgentCount = st.DirectAgentCount
		}
		items = append(items, item)
	}
	return &agentpb.TreeNodeAgentsResp{Items: items}, nil
}
