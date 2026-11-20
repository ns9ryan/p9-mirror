package audit

import "oa.98ent.com/p9/operator-activity/modules/activity/types"

// Decision 审核决策
type Decision int

const (
	DecisionPending  Decision = 1 // 人工审核
	DecisionAutoPass Decision = 2 // 自动通过
)

// Strategy 审核策略：仅决策，无副作用
type Strategy interface {
	Decide(order *types.ActivityOrder) Decision
}

// ConfigStrategy 按活动配置决策
type ConfigStrategy struct {
	AutoAudit       bool
	AutoAuditMaxAmt int64
}

func (s *ConfigStrategy) Decide(order *types.ActivityOrder) Decision {
	if s == nil || !s.AutoAudit {
		return DecisionPending
	}
	if order.RewardsTotal <= s.AutoAuditMaxAmt {
		return DecisionAutoPass
	}
	return DecisionPending
}

// BuildFromSettings 按活动配置创建审核策略
func BuildFromSettings(activity *types.Activity) Strategy {
	return &ConfigStrategy{
		AutoAudit:       activity.AutoAudit,
		AutoAuditMaxAmt: activity.AutoAuditMaxAmt,
	}
}
