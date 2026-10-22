package dao

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/membervipconfig"
)

// MemberVipConfig 会员VIP配置数据访问
type MemberVipConfig struct {
	db *ent.Client // Ent数据库客户端
}

// MemberVipConfigUpdate VIP配置更新字段
type MemberVipConfigUpdate struct {
	VipEnabled            *bool  // VIP总开关
	UpgradeRewardEnabled  *bool  // 晋级奖励开关
	RetentionCheckEnabled *bool  // 保级考核开关
	RewardGrantMode       *int64 // 奖励派发模式: 1手动领取, 2自动派发
	RewardWaterMultiplier *int64 // VIP晋级奖励洗码倍率，放大100000倍存储
}

// GetConfig 获取当前厅VIP基础配置
func (d *MemberVipConfig) GetConfig(ctx context.Context) (*ent.MemberVipConfig, error) {
	// 查询当前厅已保存的VIP配置，不存在时使用默认值
	config, err := d.db.MemberVipConfig.Query().Only(ctx)
	if err == nil {
		return config, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	// 尚未保存配置时，返回Ent Schema默认值
	return &ent.MemberVipConfig{
		VipEnabled:            membervipconfig.DefaultVipEnabled,            // VIP总开关
		UpgradeRewardEnabled:  membervipconfig.DefaultUpgradeRewardEnabled,  // 晋级奖励开关
		RetentionCheckEnabled: membervipconfig.DefaultRetentionCheckEnabled, // 保级考核开关
		RewardGrantMode:       membervipconfig.DefaultRewardGrantMode,       // 奖励派发模式: 1手动领取, 2自动派发
		RewardWaterMultiplier: membervipconfig.DefaultRewardWaterMultiplier, // VIP晋级奖励洗码倍率，放大100000倍存储
	}, nil
}

// UpdateConfig 新增或更新当前厅VIP基础配置
func (d *MemberVipConfig) UpdateConfig(ctx context.Context, in MemberVipConfigUpdate) error {
	now := time.Now().UTC()

	// 创建VIP配置: operator_code 由Ent Hook从可信上下文绑定
	create := d.db.MemberVipConfig.Create().
		SetUpdatedAt(now).                                          // 更新时间
		SetNillableVipEnabled(in.VipEnabled).                       // VIP总开关
		SetNillableUpgradeRewardEnabled(in.UpgradeRewardEnabled).   // 晋级奖励开关
		SetNillableRetentionCheckEnabled(in.RetentionCheckEnabled). // 保级考核开关
		SetNillableRewardGrantMode(in.RewardGrantMode).             // 奖励派发模式: 1手动领取, 2自动派发
		SetNillableRewardWaterMultiplier(in.RewardWaterMultiplier)  // VIP晋级奖励洗码倍率，放大100000倍存储

	// 首次开启时记录时间；关闭状态下不记录
	if in.VipEnabled != nil && *in.VipEnabled {
		create.SetLastEnabledAt(now)
	}

	// 按厅编码处理并发首次保存，未传入的字段不覆盖
	return create.OnConflictColumns(membervipconfig.FieldOperatorCode).
		Update(func(u *ent.MemberVipConfigUpsert) {
			u.SetUpdatedAt(now) // 更新时间

			if in.VipEnabled != nil {
				u.SetVipEnabled(*in.VipEnabled)

				if *in.VipEnabled {
					// 仅从关闭切换为开启时，更新最近开启时间
					u.Set(membervipconfig.FieldLastEnabledAt, sql.Expr(
						`CASE
							WHEN NOT member_vip_config.vip_enabled
							THEN excluded.last_enabled_at
							ELSE member_vip_config.last_enabled_at
						END`,
					))
				}
			}

			// 晋级奖励开关
			if in.UpgradeRewardEnabled != nil {
				u.SetUpgradeRewardEnabled(*in.UpgradeRewardEnabled)
			}
			// 保级考核开关
			if in.RetentionCheckEnabled != nil {
				u.SetRetentionCheckEnabled(*in.RetentionCheckEnabled)
			}
			// 奖励派发模式: 1手动领取, 2自动派发
			if in.RewardGrantMode != nil {
				u.SetRewardGrantMode(*in.RewardGrantMode)
			}
			// VIP晋级奖励洗码倍率，放大100000倍存储
			if in.RewardWaterMultiplier != nil {
				u.SetRewardWaterMultiplier(*in.RewardWaterMultiplier)
			}
		}).
		Exec(ctx)
}
