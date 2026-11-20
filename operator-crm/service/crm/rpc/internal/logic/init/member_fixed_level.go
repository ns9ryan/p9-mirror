package initlogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevel"
)

// initMemberFixedLevels 初始化预设固定层级, 已有层级时不重复创建
func initMemberFixedLevels(ctx context.Context, tx *ent.Tx) error {
	// 已有固定层级时跳过, 避免运营修改名称后再次创建
	exists, err := tx.MemberFixedLevel.Query().Exist(ctx)
	if err != nil {
		return fmt.Errorf("查询固定层级失败: %w", err)
	}
	if exists {
		return nil
	}

	// 预设固定层级
	levels := []struct {
		name        string // 固定层级名称
		description string // 固定层级说明
	}{
		{"套利玩家", "存在套利行为风险的会员"},
		{"KOL玩家", "具有推广影响力的会员"},
		{"邀请玩家", "通过邀请渠道加入的会员"},
		{"恶意玩家", "存在异常或恶意行为风险的会员"},
		{"大额玩家", "交易金额较大的会员"},
	}

	for _, level := range levels {
		if err := tx.MemberFixedLevel.Create().
			SetName(level.name).               // 固定层级名称
			SetDescription(level.description). // 固定层级说明
			OnConflictColumns(
				memberfixedlevel.FieldOperatorCode, // 所属厅业务编码
				memberfixedlevel.FieldName,         // 固定层级名称
			).
			DoNothing().
			Exec(ctx); err != nil {
			return fmt.Errorf("初始化固定层级 %s 失败: %w", level.name, err)
		}
	}

	return nil
}
