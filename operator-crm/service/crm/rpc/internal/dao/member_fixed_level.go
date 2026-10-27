package dao

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevel"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
)

// MemberFixedLevel 会员固定层级数据访问
type MemberFixedLevel struct {
	db *ent.Client // Ent数据库客户端
}

// Create 创建会员固定层级
func (d *MemberFixedLevel) Create(
	ctx context.Context,
	name, description string,
	withdrawalLimit *int64,
) (*ent.MemberFixedLevel, error) {
	create := d.db.MemberFixedLevel.Create().
		SetName(name).              // 固定层级名称
		SetDescription(description) // 固定层级说明

	if withdrawalLimit != nil {
		create.SetWithdrawalLimit(*withdrawalLimit) // 单笔提款金额上限
	}

	// 所属厅编码由Ent Hook从可信上下文绑定
	return create.Save(ctx)
}

// List 获取固定层级分页列表
func (d *MemberFixedLevel) List(
	ctx context.Context,
	name string,
	offset, limit int,
) ([]*ent.MemberFixedLevel, int64, error) {
	query := d.db.MemberFixedLevel.Query()

	// 按名称筛选
	if name = strings.TrimSpace(name); name != "" {
		query = query.Where(memberfixedlevel.NameContainsFold(name))
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
		Order(memberfixedlevel.ByID(sql.OrderDesc())).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return levels, int64(total), nil
}

// MemberCounts 统计指定固定层级的会员数
func (d *MemberFixedLevel) MemberCounts(
	ctx context.Context,
	levelIDs []int64,
) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(levelIDs))
	if len(levelIDs) == 0 {
		return counts, nil
	}

	// 按固定层级分组统计关联会员数
	var results []struct {
		FixedLevelID int64 `json:"fixed_level_id"` // 固定层级ID
		Count        int64 `json:"count"`          // 会员数
	}

	if err := d.db.MemberFixedLevelAssignment.Query().
		Where(memberfixedlevelassignment.FixedLevelIDIn(levelIDs...)).
		GroupBy(memberfixedlevelassignment.FieldFixedLevelID).
		Aggregate(ent.Count()).
		Scan(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		counts[result.FixedLevelID] = result.Count
	}

	return counts, nil
}

// Update 修改会员固定层级
func (d *MemberFixedLevel) Update(
	ctx context.Context,
	id int64,
	name, description string,
	withdrawalLimit *int64,
) error {
	update := d.db.MemberFixedLevel.UpdateOneID(id).
		SetName(name).              // 固定层级名称
		SetDescription(description) // 固定层级说明

	if withdrawalLimit == nil {
		update.ClearWithdrawalLimit() // 清除提款上限
	} else {
		update.SetWithdrawalLimit(*withdrawalLimit) // 单笔提款金额上限
	}

	// 更新操作由Ent Hook限定当前厅
	_, err := update.Save(ctx)
	return err
}

// Delete 删除固定层级及其会员关联
func (d *MemberFixedLevel) Delete(ctx context.Context, id int64) error {
	return WithTx(ctx, d.db, func(tx *ent.Tx) error {
		// 先确认层级存在且属于当前厅
		if _, err := tx.MemberFixedLevel.Query().
			Where(memberfixedlevel.ID(id)).
			Only(ctx); err != nil {
			return err
		}

		// 先解除所有会员关联
		if _, err := tx.MemberFixedLevelAssignment.Delete().
			Where(memberfixedlevelassignment.FixedLevelID(id)).
			Exec(ctx); err != nil {
			return err
		}

		// 再删除固定层级
		return tx.MemberFixedLevel.DeleteOneID(id).Exec(ctx)
	})
}
