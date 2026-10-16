package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentstats"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/pkg/agentline"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

const (
	defaultPageSize int32 = 20  // 默认每页20条
	maxPageSize     int32 = 100 // 最大每页100条

	timeFilterRegister  int32 = 1 // 注册时间过滤
	timeFilterLastLogin int32 = 2 // 最后登录时间过滤

	numberFilterDirectAgent  int32 = 1 // 直属代理数量过滤
	numberFilterTotalAgent   int32 = 2 // 总属理数量过滤
	numberFilterDirectMember int32 = 3 // 直属成员数量过滤
	numberFilterTotalMember  int32 = 4 // 总成员数量过滤
	numberFilterShareRate    int32 = 5 // 分享比例过滤
	numberFilterTotalComm    int32 = 6 // 总佣金过滤
	numberFilterClaimedComm  int32 = 7 // 已领取佣金过滤

	agentFilterSelf   int32 = 1 // 自己过滤
	agentFilterParent int32 = 2 // 上级过滤
)

// ListOperatorAgents 按层级查询代理，并按 agent_line_code 分组。
func (s *Service) ListOperatorAgents(ctx context.Context, in *agentpb.ListOperatorAgentsReq) (*agentpb.ListOperatorAgentsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Level < 1 || in.Level > int32(maxLevel) {
		return nil, xerr.BadRequest(i18n.AgentLevelOutOfRange)
	}

	rows, err := s.DB.AgentUser.Query().
		Where(agentuser.LevelEQ(int64(in.Level))).
		Order(agentuser.ByID(sql.OrderDesc())).
		Select(agentuser.FieldID, agentuser.FieldAccount, agentuser.FieldLevel, agentuser.FieldAgentLineCode).
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

	byLine := make(map[string][]*agentpb.AgentSearchItem, len(rows))
	unknownOrder := make([]string, 0)
	for _, row := range rows {
		item := &agentpb.AgentSearchItem{
			Id:         row.ID,
			Account:    row.Account,
			Level:      int32(row.Level),
			LevelAlias: levelAlias[row.Level],
			LineCode:   row.AgentLineCode,
		}
		line := row.AgentLineCode
		if _, ok := byLine[line]; !ok {
			if !agentline.IsValid(line) {
				unknownOrder = append(unknownOrder, line)
			}
		}
		byLine[line] = append(byLine[line], item)
	}

	groups := make([]*agentpb.OperatorAgentLineGroup, 0, len(byLine))
	for _, code := range agentline.All() {
		agents, ok := byLine[code]
		if !ok || len(agents) == 0 {
			continue
		}
		groups = append(groups, &agentpb.OperatorAgentLineGroup{
			LineCode: code,
			Agents:   agents,
		})
		delete(byLine, code)
	}
	for _, code := range unknownOrder {
		agents, ok := byLine[code]
		if !ok || len(agents) == 0 {
			continue
		}
		groups = append(groups, &agentpb.OperatorAgentLineGroup{
			LineCode: code,
			Agents:   agents,
		})
	}

	return &agentpb.ListOperatorAgentsResp{Groups: groups}, nil
}

// ListAgents 租户内分页代理列表。
func (s *Service) ListAgents(ctx context.Context, in *agentpb.ListAgentsReq) (*agentpb.ListAgentsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		in = &agentpb.ListAgentsReq{}
	}

	page := max(in.Page, 1)
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	// 构建查询条件
	preds, err := buildListPredicates(ctx, s.DB, in)
	if err != nil {
		return nil, err
	}

	q := s.DB.AgentUser.Query()
	if len(preds) > 0 {
		q = q.Where(preds...)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.
		WithStats().
		Where(agentuser.TopIDNotNil()).
		Order(agentuser.ByID(sql.OrderDesc())).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	parentIDs := make([]int64, 0, len(rows))
	topIDs := make([]int64, 0, len(rows))
	seenParent := make(map[int64]struct{}, len(rows))
	seenTop := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row.ParentID != nil && *row.ParentID > 0 {
			if _, ok := seenParent[*row.ParentID]; !ok {
				seenParent[*row.ParentID] = struct{}{}
				parentIDs = append(parentIDs, *row.ParentID)
			}
		}
		if row.TopID != nil && *row.TopID > 0 {
			if _, ok := seenTop[*row.TopID]; !ok {
				seenTop[*row.TopID] = struct{}{}
				topIDs = append(topIDs, *row.TopID)
			}
		}
	}

	accountByID := make(map[int64]string, len(parentIDs)+len(topIDs))
	lookupIDs := uniqueIDs(append(parentIDs, topIDs...))
	if len(lookupIDs) > 0 {
		refs, qerr := s.DB.AgentUser.Query().
			Where(agentuser.IDIn(lookupIDs...)).
			Select(agentuser.FieldID, agentuser.FieldAccount).
			All(ctx)
		if qerr != nil {
			return nil, qerr
		}
		for _, ref := range refs {
			accountByID[ref.ID] = ref.Account
		}
	}
	// 获取等级别名
	levelAlias := make(map[int64]string)
	levels, err := s.DB.AgentLevel.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	for _, lvl := range levels {
		levelAlias[lvl.Level] = lvl.Alias
	}

	items := make([]*agentpb.AgentListItem, 0, len(rows))
	for _, row := range rows {
		item := &agentpb.AgentListItem{
			Id:               row.ID,
			Account:          row.Account,
			Level:            int32(row.Level),
			LevelAlias:       levelAlias[row.Level],
			RegisterSource:   int32(row.RegisterSource),
			ShareRate:        int32(row.ShareRate),
			WalletBalance:    0,
			Status:           int32(row.Status),
			PayoutRestricted: int32(row.PayoutRestricted),
			PromoLink:        row.PromoCode,
		}
		if row.ParentID != nil {
			item.ParentId = *row.ParentID
			item.ParentAccount = accountByID[*row.ParentID]
		}
		if row.TopID != nil {
			item.TopId = *row.TopID
			item.TopAccount = accountByID[*row.TopID]
		}
		if row.Remark != nil {
			item.Remark = *row.Remark
		}
		if row.RegisterAt != nil {
			item.RegisterAt = row.RegisterAt.Unix()
		}
		if row.LastLoginAt != nil {
			item.LastLoginAt = row.LastLoginAt.Unix()
		}
		if st := row.Edges.Stats; st != nil {
			item.DirectMemberCount = st.DirectMemberCount
			item.DirectAgentCount = st.DirectAgentCount
			item.TotalMemberCount = st.TotalMemberCount
			item.TotalAgentCount = st.TotalAgentCount
			item.TotalCommission = st.TotalCommissionAmount
			item.UnpaidCommission = st.UnpaidCommissionAmount
			item.VisitUv = st.VisitUv
		}
		items = append(items, item)
	}

	return &agentpb.ListAgentsResp{
		Items: items,
		Total: int64(total),
	}, nil
}

// buildListPredicates 构建查询条件。
func buildListPredicates(ctx context.Context, db *ent.Client, in *agentpb.ListAgentsReq) ([]predicate.AgentUser, error) {
	preds := make([]predicate.AgentUser, 0, 8)

	if in.RegisterSource > 0 {
		preds = append(preds, agentuser.RegisterSourceEQ(int64(in.RegisterSource)))
	}
	if in.Level > 0 {
		preds = append(preds, agentuser.LevelEQ(int64(in.Level)))
	}
	if in.Status > 0 {
		preds = append(preds, agentuser.StatusEQ(int64(in.Status)))
	}
	if in.PayoutRestricted > 0 {
		preds = append(preds, agentuser.PayoutRestrictedEQ(int64(in.PayoutRestricted)))
	}

	if in.TimeFilterType > 0 && in.TimeStart > 0 && in.TimeEnd > 0 {
		start := time.Unix(in.TimeStart, 0)
		end := time.Unix(in.TimeEnd, 0)
		switch in.TimeFilterType {
		case timeFilterRegister:
			preds = append(preds, agentuser.RegisterAtGTE(start), agentuser.RegisterAtLTE(end))
		case timeFilterLastLogin:
			preds = append(preds, agentuser.LastLoginAtGTE(start), agentuser.LastLoginAtLTE(end))
		}
	}

	if p := numberFilterPredicate(in); p != nil {
		preds = append(preds, p)
	}

	agentPred, err := agentKeywordPredicate(ctx, db, in)
	if err != nil {
		return nil, err
	}
	if agentPred != nil {
		preds = append(preds, agentPred)
	}

	return preds, nil
}

// numberFilterPredicate 构建数量过滤条件。
func numberFilterPredicate(in *agentpb.ListAgentsReq) predicate.AgentUser {
	if in.NumberFilterType <= 0 || (in.NumberMin <= 0 && in.NumberMax <= 0) {
		return nil
	}
	minV, maxV := in.NumberMin, in.NumberMax

	switch in.NumberFilterType {
	case numberFilterShareRate:
		ps := make([]predicate.AgentUser, 0, 2)
		if minV > 0 {
			ps = append(ps, agentuser.ShareRateGTE(minV))
		}
		if maxV > 0 {
			ps = append(ps, agentuser.ShareRateLTE(maxV))
		}
		if len(ps) == 1 {
			return ps[0]
		}
		return agentuser.And(ps...)
	case numberFilterDirectMember:
		return agentuser.HasStatsWith(statsRangePreds(agentstats.FieldDirectMemberCount, minV, maxV)...)
	case numberFilterDirectAgent:
		return agentuser.HasStatsWith(statsRangePreds(agentstats.FieldDirectAgentCount, minV, maxV)...)
	case numberFilterTotalAgent:
		return agentuser.HasStatsWith(statsRangePreds(agentstats.FieldTotalAgentCount, minV, maxV)...)
	case numberFilterTotalMember:
		return agentuser.HasStatsWith(statsRangePreds(agentstats.FieldTotalMemberCount, minV, maxV)...)
	case numberFilterTotalComm:
		return agentuser.HasStatsWith(statsRangePreds(agentstats.FieldTotalCommissionAmount, minV, maxV)...)
	case numberFilterClaimedComm:
		return agentuser.HasStatsWith(claimedCommissionPred(minV, maxV))
	default:
		return nil
	}
}

// statsRangePreds 构建统计范围过滤条件。
func statsRangePreds(field string, minV, maxV int64) []predicate.AgentStats {
	ps := make([]predicate.AgentStats, 0, 2)
	if minV > 0 {
		ps = append(ps, predicate.AgentStats(func(s *sql.Selector) {
			s.Where(sql.GTE(s.C(field), minV))
		}))
	}
	if maxV > 0 {
		ps = append(ps, predicate.AgentStats(func(s *sql.Selector) {
			s.Where(sql.LTE(s.C(field), maxV))
		}))
	}
	return ps
}

// claimedCommissionPred 构建已领取佣金范围过滤条件。
func claimedCommissionPred(minV, maxV int64) predicate.AgentStats {
	return func(s *sql.Selector) {
		expr := fmt.Sprintf("(%s - %s)", s.C(agentstats.FieldTotalCommissionAmount), s.C(agentstats.FieldUnpaidCommissionAmount))
		if minV > 0 {
			s.Where(sql.ExprP(expr+" >= ?", minV))
		}
		if maxV > 0 {
			s.Where(sql.ExprP(expr+" <= ?", maxV))
		}
	}
}

// agentKeywordPredicate 构建代理关键词过滤条件。
func agentKeywordPredicate(ctx context.Context, db *ent.Client, in *agentpb.ListAgentsReq) (predicate.AgentUser, error) {
	kw := strings.TrimSpace(in.AgentKeyword)
	if in.AgentFilterType <= 0 || kw == "" {
		return nil, nil
	}

	id, isID := dao.ParsePureDigitsID(kw)

	switch in.AgentFilterType {
	case agentFilterSelf:
		if isID {
			return agentuser.IDEQ(id), nil
		}
		return agentuser.AccountContainsFold(kw), nil
	case agentFilterParent:
		if isID {
			return agentuser.ParentIDEQ(id), nil
		}
		parents, err := db.AgentUser.Query().
			Where(agentuser.AccountContainsFold(kw)).
			Select(agentuser.FieldID).
			All(ctx)
		if err != nil {
			return nil, err
		}
		if len(parents) == 0 {
			return agentuser.IDEQ(0), nil // no match
		}
		ids := make([]int64, 0, len(parents))
		for _, p := range parents {
			ids = append(ids, p.ID)
		}
		return agentuser.ParentIDIn(ids...), nil
	default:
		return nil, nil
	}
}

// uniqueIDs 去重ID。
func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
