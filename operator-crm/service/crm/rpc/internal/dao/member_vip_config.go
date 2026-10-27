package dao

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-crm/ent"
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
	RewardWaterMultiplier *int64 // VIP晋级奖励洗码倍率, 放大100000倍存储
}

// GetConfig 获取当前厅VIP基础配置
func (d *MemberVipConfig) GetConfig(ctx context.Context) (*ent.MemberVipConfig, error) {
	return d.db.MemberVipConfig.Query().Only(ctx)
}

// UpdateConfig 更新当前厅已初始化的VIP基础配置
func (d *MemberVipConfig) UpdateConfig(ctx context.Context, in MemberVipConfigUpdate) error {
	// 查询已有配置, 不存在时直接返回错误
	config, err := d.GetConfig(ctx)
	if err != nil {
		return err
	}

	// 只更新本次传入的字段
	update := d.db.MemberVipConfig.UpdateOneID(config.ID).
		SetNillableVipEnabled(in.VipEnabled).                       // VIP总开关
		SetNillableUpgradeRewardEnabled(in.UpgradeRewardEnabled).   // 晋级奖励开关
		SetNillableRetentionCheckEnabled(in.RetentionCheckEnabled). // 保级考核开关
		SetNillableRewardGrantMode(in.RewardGrantMode).             // 奖励派发模式
		SetNillableRewardWaterMultiplier(in.RewardWaterMultiplier)  // VIP晋级奖励洗码倍率

	// 仅从关闭切换为开启时, 更新最近开启时间
	if in.VipEnabled != nil && *in.VipEnabled && !config.VipEnabled {
		update.SetLastEnabledAt(time.Now())
	}

	return update.Exec(ctx)
}
