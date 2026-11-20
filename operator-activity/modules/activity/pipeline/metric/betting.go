package metric

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// BettingCalculator 投注指标（直接用循环周期管理器取当前周期窗）
type BettingCalculator struct {
	Store MetricStore
	Cycle *pipeline.Manager
}

func (c *BettingCalculator) CalcMetric(ctx context.Context, activity *types.Activity, evt types.ActivityEvent) (*types.TriggerMetric, error) {
	var (
		betCount  int64
		betAmount int64
		err       error
	)
	_, window, _ := c.Cycle.CurrentRound(activity, evt.GetEventTime())
	// 累计领取模式，需要累计投注金额
	if activity.Settings.BetClaimMode == types.BetClaimModeByCumulative {
		betAmount, err = c.Store.SumBets(ctx, evt.GetMemberID(), window)
		if err != nil {
			return nil, err
		}
	}

	// betCount, err = c.Store.CountBets(evt.UserID, window)
	// if err != nil {
	// 	return nil, err
	// }

	return &types.TriggerMetric{
		BetAmount:     evt.GetAmount(),
		BetCount:      betCount,
		CumulativeBet: betAmount,
	}, nil
}
