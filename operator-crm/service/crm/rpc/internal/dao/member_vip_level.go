package dao

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberviplevel"
)

// MemberVipLevel 会员VIP等级数据访问
type MemberVipLevel struct {
	db *ent.Client // Ent数据库客户端
}

// MemberVipLevelCreate VIP等级创建参数
type MemberVipLevelCreate struct {
	UpgradeDeposit   int64 // 晋级所需存款金额
	UpgradeWater     int64 // 晋级所需洗码量
	RetentionDeposit int64 // 保级所需存款金额
	RetentionWater   int64 // 保级所需洗码量
	UpgradeReward    int64 // 晋级奖励金额
}

// MemberVipLevelUpdate VIP等级更新字段
type MemberVipLevelUpdate struct {
	UpgradeDeposit   *int64 // 晋级所需存款金额
	UpgradeWater     *int64 // 晋级所需洗码量
	RetentionDeposit *int64 // 保级所需存款金额
	RetentionWater   *int64 // 保级所需洗码量
	UpgradeReward    *int64 // 晋级奖励金额
}

// List 获取VIP等级分页列表
func (d *MemberVipLevel) List(ctx context.Context, offset, limit int) ([]*ent.MemberVipLevel, int64, error) {
	query := d.db.MemberVipLevel.Query()

	// 统计等级总数
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 按等级编号升序查询
	levels, err := query.
		Order(memberviplevel.ByLevel(sql.OrderAsc())).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return levels, int64(total), nil
}

// Get 根据ID获取VIP等级
func (d *MemberVipLevel) Get(ctx context.Context, id int64) (*ent.MemberVipLevel, error) {
	return d.db.MemberVipLevel.Query().
		Where(memberviplevel.ID(id)).
		Only(ctx)
}

// GetDefaultTx 获取默认VIP等级
func (d *MemberVipLevel) GetDefaultTx(ctx context.Context, tx *ent.Tx) (*ent.MemberVipLevel, error) {
	return tx.MemberVipLevel.Query().
		Where(memberviplevel.LevelEQ(0)). // VIP等级编号, 从0开始连续递增
		Only(ctx)
}

// GetHighest 获取当前最高VIP等级
func (d *MemberVipLevel) GetHighest(ctx context.Context) (*ent.MemberVipLevel, error) {
	return d.db.MemberVipLevel.Query().
		Order(memberviplevel.ByLevel(sql.OrderDesc())).
		First(ctx)
}

// Create 创建VIP等级并分配下一个等级编号
func (d *MemberVipLevel) Create(ctx context.Context, in MemberVipLevelCreate) (*ent.MemberVipLevel, error) {
	// 初始VIP0必须由分站初始化流程创建
	last, err := d.GetHighest(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("VIP等级初始数据不存在, 请检查分站初始化状态: %w", err)
		}
		return nil, fmt.Errorf("查询最高VIP等级失败: %w", err)
	}

	// 等级编号从当前最高等级递增
	return d.db.MemberVipLevel.Create().
		SetLevel(last.Level + 1).                 // VIP等级编号, 从0开始连续递增
		SetUpgradeDeposit(in.UpgradeDeposit).     // 晋级所需存款金额
		SetUpgradeWater(in.UpgradeWater).         // 晋级所需洗码量
		SetRetentionDeposit(in.RetentionDeposit). // 保级所需存款金额
		SetRetentionWater(in.RetentionWater).     // 保级所需洗码量
		SetUpgradeReward(in.UpgradeReward).       // 晋级奖励金额
		Save(ctx)
}

// Update 修改VIP等级配置
func (d *MemberVipLevel) Update(ctx context.Context, id int64, in MemberVipLevelUpdate) error {
	return d.db.MemberVipLevel.UpdateOneID(id).
		SetNillableUpgradeDeposit(in.UpgradeDeposit).     // 晋级所需存款金额
		SetNillableUpgradeWater(in.UpgradeWater).         // 晋级所需洗码量
		SetNillableRetentionDeposit(in.RetentionDeposit). // 保级所需存款金额
		SetNillableRetentionWater(in.RetentionWater).     // 保级所需洗码量
		SetNillableUpgradeReward(in.UpgradeReward).       // 晋级奖励金额
		Exec(ctx)
}

// Delete 删除指定VIP等级
func (d *MemberVipLevel) Delete(ctx context.Context, id int64) error {
	return d.db.MemberVipLevel.DeleteOneID(id).Exec(ctx)
}

// MemberCounts 统计指定VIP等级的会员数
func (d *MemberVipLevel) MemberCounts(ctx context.Context, levelIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(levelIDs))
	if len(levelIDs) == 0 {
		return counts, nil
	}

	// 按VIP等级分组统计会员数
	var results []struct {
		VipLevelID int64 `json:"vip_level_id"` // VIP等级ID
		Count      int64 `json:"count"`        // 会员数
	}

	err := d.db.Member.Query().
		Where(member.VipLevelIDIn(levelIDs...)).
		GroupBy(member.FieldVipLevelID).
		Aggregate(ent.Count()).
		Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		counts[result.VipLevelID] = result.Count
	}

	return counts, nil
}
