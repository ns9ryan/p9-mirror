package reward

import "oa.98ent.com/p9/operator-activity/modules/activity/types"

// BuildFromSettings 按活动奖励模式创建计算器
func BuildFromSettings(activity *types.Activity) Calculator {
	base := BaseCalculator{Schemas: activity.Settings.RewardSchemas}
	switch activity.Settings.BonusMode {
	case types.BonusFixed:
		return &FixedCalculator{BaseCalculator: base}
	case types.BonusPercentage:
		return &PercentCalculator{BaseCalculator: base}
	case types.BonusFixedTiered:
		return &FixedTieredCalculator{BaseCalculator: base}
	case types.BonusPercentageTiered:
		return &PercentageTieredCalculator{BaseCalculator: base}
	case types.BonusSeparate:
		return &SeparateCalculator{BaseCalculator: base}
	default:
		return &FixedCalculator{BaseCalculator: base}
	}
}
