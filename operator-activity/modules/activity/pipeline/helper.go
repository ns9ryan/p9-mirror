package pipeline

import (
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// wheelUseDepositMetric 当前事件是否按累计充值判档（投注|充值双开时随事件切换）
func wheelUseDepositMetric(ctx *EvalContext) bool {
	return ctx != nil && ctx.Metric != nil && ctx.Metric.FromDeposit
}

// filterThreshold 过滤达标档（本次事件前已达标则跳过，避免重复发放）
func FilterThreshold(ctx *EvalContext, schema types.RewardSchema) bool {
	// if ctx.Promo.Type == types.PromoTypeDeposit {
	// 	switch ctx.Promo.Settings.DepositRewardType {
	// 	case types.DepositCumulative:
	// 		return ctx.Metric.CumulativeDeposit-ctx.Metric.DepositAmount >= schema.Threshold
	// 	}
	// }
	// if ctx.Promo.Type == types.PromoTypeLuckyWheel {
	// 	if wheelUseDepositMetric(ctx) {
	// 		return ctx.Metric.CumulativeDeposit-ctx.Metric.DepositAmount >= schema.Threshold
	// 	}
	// 	return ctx.Metric.CumulativeBet-ctx.Metric.BetAmount >= schema.Threshold
	// }
	return false
}

// matchThreshold 匹配达标档
func MatchThreshold(ctx *EvalContext, schema types.RewardSchema) bool {
	if ctx.Activity.Type == types.ActivityTypeReferralMember {
		return int64(ctx.Metric.ValidReferralCount) >= int64(schema.Count)
	}
	return MetricValue(ctx) >= schema.Threshold
}

// metricValue 计算指标值
func MetricValue(ctx *EvalContext) int64 {
	if ctx.Metric == nil {
		return 0
	}
	switch ctx.Activity.Type {
	case types.ActivityTypeReferralMember:
		if ctx.Activity.Settings.BonusMode == types.BonusSeparate {
			return int64(ctx.Metric.NewVaildReferralCount)
		}
		return int64(ctx.Metric.ValidReferralCount)
	case types.ActivityTypeWagering:
		if ctx.Activity.Settings.BetClaimMode == types.BetClaimModeByCumulative {
			return ctx.Metric.CumulativeBet
		}
		return ctx.Metric.BetAmount
	case types.ActivityTypeDeposit:
		if ctx.Activity.Settings.DepositRewardType == types.DepositCumulative {
			return ctx.Metric.CumulativeDeposit
		}
		return ctx.Metric.DepositAmount
	case types.ActivityTypeLuckyWheel:
		if wheelUseDepositMetric(ctx) {
			return ctx.Metric.CumulativeDeposit
		}
		return ctx.Metric.CumulativeBet
	default:
		if ctx.Metric.CumulativeDeposit > 0 {
			return ctx.Metric.CumulativeDeposit
		}
		return ctx.Metric.DepositAmount
	}
}
