package dao

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/member"
	entlevel "oa.98ent.com/p9/operator-crm/ent/memberautolevel"
)

// MemberAutoLevel 会员自动层级数据访问
type MemberAutoLevel struct {
	db *ent.Client // Ent数据库客户端
}

// Create 创建会员自动层级
func (d *MemberAutoLevel) Create(ctx context.Context, name string, minDeposit, minWater int64) (*ent.MemberAutoLevel, error) {
	return d.db.MemberAutoLevel.Create().
		SetName(name).             // 自动层级名称
		SetMinDeposit(minDeposit). // 晋级要求的累计存款金额
		SetMinWater(minWater).     // 晋级要求的累计洗码量
		Save(ctx)
}

// HasThresholdConflict 检查晋级条件是否与已有层级冲突
func (d *MemberAutoLevel) HasThresholdConflict(ctx context.Context, minDeposit, minWater int64) (bool, error) {
	// 两项条件必须分别严格大于当前厅所有已有层级
	return d.db.MemberAutoLevel.Query().
		Where(entlevel.Or(
			entlevel.MinDepositGTE(minDeposit), // 晋级要求的累计存款金额
			entlevel.MinWaterGTE(minWater),     // 晋级要求的累计洗码量
		)).
		Exist(ctx)
}

// List 获取自动层级分页列表
func (d *MemberAutoLevel) List(ctx context.Context, name string, offset, limit int) ([]*ent.MemberAutoLevel, int64, error) {
	query := d.db.MemberAutoLevel.Query()

	// 按名称筛选
	if name = strings.TrimSpace(name); name != "" {
		query = query.Where(entlevel.NameContainsFold(name))
	}

	// 统计符合条件的层级数
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

// MemberCounts 统计指定自动层级的会员数
func (d *MemberAutoLevel) MemberCounts(ctx context.Context, levelIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(levelIDs))
	if len(levelIDs) == 0 {
		return counts, nil
	}

	// 按层级分组统计
	var results []struct {
		AutoLevelID int64 `json:"auto_level_id"` // 自动层级ID
		Count       int64 `json:"count"`         // 会员数
	}

	err := d.db.Member.Query().
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
