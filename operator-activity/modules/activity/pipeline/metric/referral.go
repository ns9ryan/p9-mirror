package metric

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// ReferralCalculator 推荐指标
type ReferralCalculator struct {
	Store MetricStore
	Cycle *pipeline.Manager
}

func (c *ReferralCalculator) CalcMetric(ctx context.Context, activity *types.Activity, evt types.ActivityEvent) (*types.TriggerMetric, error) {
	_, window, _ := c.Cycle.CurrentRound(activity, evt.GetEventTime())

	totalCount, validCount, newCount, err := c.Store.CountValidReferrals(ctx, evt, activity, window)
	if err != nil {
		return nil, err
	}

	return &types.TriggerMetric{
		TotalReferralCount:    totalCount,
		ValidReferralCount:    validCount,
		NewVaildReferralCount: newCount,
	}, nil
}
