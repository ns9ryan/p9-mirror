package agent

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

const (
	defaultSearchLevel int32 = 2
	defaultSearchLimit int32 = 20
	maxSearchLimit     int32 = 50
)

// SearchAgents 按 ID/账号关键字检索代理；level<=0 时默认限定 2 级。
func (s *Service) SearchAgents(ctx context.Context, in *agentpb.SearchAgentsReq) (*agentpb.SearchAgentsResp, error) {
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

	level := in.Level
	if level <= 0 {
		level = defaultSearchLevel
	}

	limit := in.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	preds := []predicate.AgentUser{
		searchKeywordPred(kw),
		agentuser.LevelGTE(int64(level)),
	}

	rows, err := s.DB.AgentUser.Query().
		Where(preds...).
		Order(agentuser.ByID(sql.OrderDesc())).
		Limit(int(limit)).
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

	items := make([]*agentpb.AgentSearchItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &agentpb.AgentSearchItem{
			Id:         row.ID,
			Account:    row.Account,
			Level:      int32(row.Level),
			LevelAlias: levelAlias[row.Level],
			LineCode:   row.AgentLineCode,
		})
	}
	return &agentpb.SearchAgentsResp{Items: items}, nil
}

func searchKeywordPred(kw string) predicate.AgentUser {
	if id, ok := dao.ParsePureDigitsID(kw); ok {
		return agentuser.Or(agentuser.IDEQ(id), agentuser.AccountContainsFold(kw))
	}
	return agentuser.AccountContainsFold(kw)
}
