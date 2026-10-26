package initlogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/membervipconfig"
	"oa.98ent.com/p9/operator-crm/ent/membervipicongroup"
	"oa.98ent.com/p9/operator-crm/ent/memberviplevel"
)

// initMemberVip 初始化VIP基础配置、VIP0及默认图标组
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

	// 默认图标每10个VIP等级共用一组, 从VIP0开始
	iconURLs := []string{
		"https://file.98ent.com/avatar/5afb77abe42d79378bb6be95ffec1a3d.png", // VIP0-9
		"https://file.98ent.com/avatar/c8bf063626f6d8e45751ce866590f97f.png", // VIP10-19
		"https://file.98ent.com/avatar/bb81b3fddfa4d802b9d4e675bade638e.png", // VIP20-29
		"https://file.98ent.com/avatar/5c6e7ff8673b6b5d30e67615c90b7527.png", // VIP30-39
		"https://file.98ent.com/avatar/f530bae5a30150d3f0dbeb82b4aec9fe.png", // VIP40-49
		"https://file.98ent.com/avatar/22eeec8f2579b1f59cfee516f3233914.png", // VIP50-59
		"https://file.98ent.com/avatar/e5e51f03020faa2f0ec273d2b49a90b1.png", // VIP60-69
		"https://file.98ent.com/avatar/c7a81a2ea4303262f88fcd833d45f9dc.png", // VIP70-79
		"https://file.98ent.com/avatar/dc20dd718a7516d3c8b54a6cf3f4f913.png", // VIP80-89
		"https://file.98ent.com/avatar/53fbe7099751f1f3bbce70689ce619d3.png", // VIP90-99
	}

	// 初始化全部默认图标组, 已存在时不覆盖运营配置
	for groupNo, iconURL := range iconURLs {
		if err := tx.MemberVipIconGroup.Create().
			SetGroupNo(int64(groupNo)). // 图标组编号, 从0开始
			SetDefaultIconURL(iconURL). // 系统默认图标URL
			SetIconURL(iconURL).        // 当前使用的图标URL
			OnConflictColumns(
				membervipicongroup.FieldOperatorCode, // 所属厅业务编码
				membervipicongroup.FieldGroupNo,      // 图标组编号, 每10个VIP等级共用一组
			).
			DoNothing().
			Exec(ctx); err != nil {
			return fmt.Errorf("初始化VIP图标组 %d 失败: %w", groupNo, err)
		}
	}

	return nil
}
