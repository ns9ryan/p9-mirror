package payout

import "oa.98ent.com/p9/operator-activity/modules/activity/types"

// BuildFromSettings 按活动派发配置创建策略
func BuildFromSettings(activity *types.Activity, store ScheduleStore, publisher PayoutPublisher, workflows WorkflowStarter) Strategy {
	var timing TimingStrategy
	switch activity.PayoutTiming {
	case types.PayoutNextDay:
		timing = &NextDayTiming{}
	default:
		timing = &RealtimeTiming{}
	}

	switch activity.PayoutMode {
	case types.PayoutManualExpireVoid:
		return &ManualClaimPayout{
			Timing:       timing,
			Store:        store,
			Publisher:    publisher,
			Workflows:    workflows,
			ExpireSec:    activity.ClaimExpireSec,
			ExpireAction: ExpireVoid,
		}
	case types.PayoutManualExpireAuto:
		return &ManualClaimPayout{
			Timing:       timing,
			Store:        store,
			Publisher:    publisher,
			Workflows:    workflows,
			ExpireSec:    activity.ClaimExpireSec,
			ExpireAction: ExpireAuto,
		}
	default:
		return &AutoPayout{
			Timing:    timing,
			Store:     store,
			Publisher: publisher,
			Workflows: workflows,
		}
	}
}
