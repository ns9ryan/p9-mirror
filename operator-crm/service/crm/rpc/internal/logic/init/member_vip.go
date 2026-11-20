package initlogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/membervipconfig"
	"oa.98ent.com/p9/operator-crm/ent/membervipicongroup"
	"oa.98ent.com/p9/operator-crm/ent/memberviplevel"
)

// initMemberVip 初始化VIP基础配置、VIP0及十组空图片记录，已有配置不覆盖。
func initMemberVip(ctx context.Context, tx *ent.Tx) error {
	// 初始化VIP基础配置, 已存在时不覆盖
	if err := tx.MemberVipConfig.Create().
		OnConflictColumns(membervipconfig.FieldOperatorCode).
		DoNothing().
		Exec(ctx); err != nil {
		return fmt.Errorf("初始化VIP基础配置失败: %w", err)
	}

	// 初始化VIP0等级, 已存在时不覆盖
	if err := tx.MemberVipLevel.Create().
		SetLevel(0).            // VIP等级编号, 从0开始连续递增
		SetUpgradeDeposit(0).   // 晋级所需存款金额
		SetUpgradeWater(0).     // 晋级所需洗码量
		SetRetentionDeposit(0). // 保级所需存款金额
		SetRetentionWater(0).   // 保级所需洗码量
		SetUpgradeReward(0).    // 晋级奖励金额
		OnConflictColumns(
			memberviplevel.FieldOperatorCode, // 所属厅业务编码
			memberviplevel.FieldLevel,        // VIP等级编号, 从0开始连续递增
		).
		DoNothing().
		Exec(ctx); err != nil {
		return fmt.Errorf("初始化VIP0等级失败: %w", err)
	}

	// 新站创建0~9共十组，图片为空表示查询时使用总站实时默认图；不调用总站RPC。
	// 继续使用外层初始化事务，任一组失败由调用方回滚；已有组不覆盖当前图片。
	for groupNo := int64(0); groupNo < 10; groupNo++ {
		if err := tx.MemberVipIconGroup.Create().
			SetGroupNo(groupNo). // 图标组编号，0~9分别对应VIP0~9至90~99。
			SetIconURL("").      // 初始化不保存固定URL；空值在查询响应中回退总站默认图。
			OnConflictColumns(
				membervipicongroup.FieldOperatorCode, // 所属厅业务编码
				membervipicongroup.FieldGroupNo,      // 图标组编号, 每10个VIP等级共用一组
			).
			DoNothing(). // 已有本站本组记录保持原值，不清空已保存图片。
			Exec(ctx); err != nil {
			return fmt.Errorf("初始化VIP图标组 %d 失败: %w", groupNo, err)
		}
	}

	return nil
}
