package agent

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentlevel"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

// ListLevels 返回当前租户代理等级列表。
func (s *Service) ListLevels(ctx context.Context, _ *agentpb.ListAgentLevelsReq) (*agentpb.ListAgentLevelsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.AgentLevel.Query().
		Order(agentlevel.ByLevel(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*agentpb.AgentLevelItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &agentpb.AgentLevelItem{
			Id:    row.ID,
			Level: int32(row.Level),
			Name:  row.Name,
			Alias: row.Alias,
		})
	}
	return &agentpb.ListAgentLevelsResp{Items: items}, nil
}

// UpdateLevels 批量 upsert 代理等级。
func (s *Service) UpdateLevels(ctx context.Context, in *agentpb.UpdateAgentLevelsReq) (*agentpb.UpdateAgentLevelsResp, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || len(in.Items) == 0 {
		return nil, xerr.BadRequest(i18n.AgentLevelItemsRequired)
	}

	seen := make(map[int32]struct{}, len(in.Items))
	items := make([]*agentpb.AgentLevelItem, 0, len(in.Items))
	for _, it := range in.Items {
		if it == nil {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		if it.Level < 1 || it.Level > int32(maxLevel) {
			return nil, xerr.BadRequest(i18n.AgentLevelOutOfRange)
		}
		name := strings.TrimSpace(it.Name)
		alias := strings.TrimSpace(it.Alias)
		if name == "" || alias == "" {
			return nil, xerr.BadRequest(i18n.InvalidParam)
		}
		if _, ok := seen[it.Level]; ok {
			return nil, xerr.BadRequest(i18n.AgentLevelDuplicate)
		}
		seen[it.Level] = struct{}{}
		items = append(items, &agentpb.AgentLevelItem{
			Level: it.Level,
			Name:  name,
			Alias: alias,
		})
	}

	out := make([]*agentpb.AgentLevelItem, 0, len(items))
	err = dao.WithTx(ctx, s.DB, func(tx *ent.Tx) error {
		for _, it := range items {
			lvl := int64(it.Level)
			row, qerr := tx.AgentLevel.Query().Where(agentlevel.LevelEQ(lvl)).Only(ctx)
			switch {
			case qerr == nil:
				row, qerr = tx.AgentLevel.UpdateOneID(row.ID).
					// SetName(it.Name).
					SetAlias(it.Alias).
					Save(ctx)
				if qerr != nil {
					return mapEntErr(qerr)
				}
			case ent.IsNotFound(qerr):
				row, qerr = tx.AgentLevel.Create().
					SetLevel(lvl).
					SetName(it.Name).
					SetAlias(it.Alias).
					Save(ctx)
				if qerr != nil {
					return mapEntErr(qerr)
				}
			default:
				return qerr
			}
			out = append(out, &agentpb.AgentLevelItem{
				Id:    row.ID,
				Level: int32(row.Level),
				Name:  row.Name,
				Alias: row.Alias,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &agentpb.UpdateAgentLevelsResp{Items: out}, nil
}

// 获取最大代理层级
func (s *Service) GetMaxLevel(ctx context.Context) (int64, error) {
	ctx, _, err := requireOperator(ctx)
	if err != nil {
		return 0, err
	}
	row, err := s.DB.AgentLevel.Query().Order(agentlevel.ByLevel(sql.OrderDesc())).First(ctx)
	if err != nil {
		return 0, err
	}
	return row.Level, nil
}
