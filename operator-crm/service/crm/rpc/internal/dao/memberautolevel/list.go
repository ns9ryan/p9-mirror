package memberautolevel

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/member"
	entlevel "oa.98ent.com/p9/operator-crm/ent/memberautolevel"
)

// List 获取自动层级分页列表
func (s *Service) List(ctx context.Context, name string, offset, limit int) ([]*ent.MemberAutoLevel, int64, error) {
	// 按层级名称筛选
	query := s.DB.MemberAutoLevel.Query()
	if name = strings.TrimSpace(name); name != "" {
		query = query.Where(entlevel.NameContainsFold(name))
	}

	// 查询数据总数
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 查询当前页数据
	levels, err := query.
		Order(entlevel.ByID(sql.OrderDesc())).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return levels, int64(total), nil
}

// CountMembersByLevelIDs 批量统计各自动层级的会员数
func (s *Service) CountMembersByLevelIDs(ctx context.Context, levelIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(levelIDs))
	if len(levelIDs) == 0 {
		return counts, nil
	}

	// 按自动层级分组统计
	var results []struct {
		AutoLevelID int64 `json:"auto_level_id"` // 自动层级ID
		Count       int64 `json:"count"`         // 会员数量
	}

	err := s.DB.Member.Query().
		Where(member.AutoLevelIDIn(levelIDs...)).
		GroupBy(member.FieldAutoLevelID).
		Aggregate(ent.Count()).
		Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		counts[result.AutoLevelID] = result.Count
	}

	return counts, nil
}
