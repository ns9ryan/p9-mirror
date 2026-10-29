package agent

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
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

	list := make([]*agentpb.TreeNodeAgentItem, 0, len(rows))
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
		list = append(list, item)
	}
	return &agentpb.TreeNodeAgentsResp{List: list}, nil
}

const maxAncestorNodes = 8

// TreeNodeAncestorAgents 按 ID 或账号精确匹配，返回当前代理及其全部上级。
func (s *Service) TreeNodeAncestorAgents(ctx context.Context, in *agentpb.TreeNodeAncestorAgentsReq) (*agentpb.TreeNodeAncestorAgentsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	kw := strings.TrimSpace(in.Keyword)
	if kw == "" {
		return nil, xerr.BadRequest(i18n.InvalidParam)
	}
	row, err := s.findExactAgent(ctx, kw)
	if err != nil {
		return nil, err
	}
	chain, err := s.ancestorChain(ctx, row)
	if err != nil {
		return nil, err
	}
	levelAlias, err := s.levelAliasMap(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*agentpb.AgentSearchItem, 0, len(chain))
	for _, node := range chain {
		list = append(list, searchItem(node, levelAlias))
	}
	return &agentpb.TreeNodeAncestorAgentsResp{List: list}, nil
}

func (s *Service) findExactAgent(ctx context.Context, kw string) (*ent.AgentUser, error) {
	agentUserFields := []string{
		agentuser.FieldID, agentuser.FieldAccount,
		agentuser.FieldLevel, agentuser.FieldAgentLineCode,
		agentuser.FieldShareRate, agentuser.FieldParentID,
	}
	// 如果kw是纯数字，则认为是代理ID
	if id, ok := dao.ParsePureDigitsID(kw); ok {
		row, err := s.FecthAgentByID(ctx, id, agentUserFields...)
		if err == nil {
			return row, nil
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
	}
	// 获取代理用户及其Profile
	row, err := s.FecthAgentByAccount(ctx, kw, agentUserFields...)
	if ent.IsNotFound(err) {
		return nil, xerr.NotFound(i18n.AgentNotFound)
	}
	return row, err
}

func (s *Service) ancestorChain(ctx context.Context, start *ent.AgentUser) ([]*ent.AgentUser, error) {
	chain := make([]*ent.AgentUser, 0, maxAncestorNodes)
	seen := make(map[int64]struct{}, maxAncestorNodes)
	cur := start
	// 获取agent_user表字段列表
	agentUserFields := []string{
		agentuser.FieldID, agentuser.FieldAccount,
		agentuser.FieldLevel, agentuser.FieldAgentLineCode,
		agentuser.FieldShareRate, agentuser.FieldParentID,
	}
	for len(chain) < maxAncestorNodes && cur != nil {
		if _, ok := seen[cur.ID]; ok {
			break
		}
		seen[cur.ID] = struct{}{}
		chain = append(chain, cur)
		if cur.ParentID == nil || *cur.ParentID <= 0 {
			break
		}
		// 获取上级代理用户
		parent, err := s.FecthAgentByID(ctx, *cur.ParentID, agentUserFields...)
		if ent.IsNotFound(err) {
			break
		}
		if err != nil {
			return nil, err
		}
		cur = parent
	}
	return chain, nil
}

func (s *Service) levelAliasMap(ctx context.Context) (map[int64]string, error) {
	levels, err := s.DB.AgentLevel.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]string, len(levels))
	for _, lvl := range levels {
		out[lvl.Level] = lvl.Alias
	}
	return out, nil
}

func searchItem(row *ent.AgentUser, levelAlias map[int64]string) *agentpb.AgentSearchItem {
	return &agentpb.AgentSearchItem{
		Id:         row.ID,
		Account:    row.Account,
		Level:      int32(row.Level),
		LevelAlias: levelAlias[row.Level],
		LineCode:   row.AgentLineCode,
		ShareRate:  int32(row.ShareRate),
	}
}
