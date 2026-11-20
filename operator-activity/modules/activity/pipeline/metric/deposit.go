package metric

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// DepositCalculator 充值指标
// 窗口内次数/金额优先走 Redis 累计计数器；key 不存在时回源 DB。
type DepositCalculator struct {
	Store   MetricStore
	Cycle   *pipeline.Manager
	Counter DepositWindowCounter
}

func (c *DepositCalculator) CalcMetric(ctx context.Context, activity *types.Activity, evt types.ActivityEvent) (*types.TriggerMetric, error) {
	switch activity.Settings.DepositRewardType {
	case types.DepositFirst:
		return &types.TriggerMetric{
			DepositCount:      evt.GetDepositCount(),
			DepositAmount:     evt.GetAmount(),
			CumulativeDeposit: evt.GetAmount(),
		}, nil
	case types.DepositSecond:
		return &types.TriggerMetric{
			DepositCount:      evt.GetDepositCount(),
			DepositAmount:     evt.GetAmount(),
			CumulativeDeposit: evt.GetAmount(),
		}, nil
	}

	_, window, _ := c.Cycle.CurrentRound(activity, evt.GetEventTime())

	count, cumulative, err := c.windowStats(ctx, activity, evt, window)
	if err != nil {
		return nil, err
	}
	return &types.TriggerMetric{
		DepositCount:      count,
		DepositAmount:     evt.GetAmount(),
		CumulativeDeposit: cumulative,
	}, nil
}

func (c *DepositCalculator) windowStats(
	ctx context.Context,
	activity *types.Activity,
	evt types.ActivityEvent,
	window types.TimeWindow,
) (count int, cumulative int64, err error) {
	load := func(ctx context.Context) (int, int64, error) {
		n, err := c.Store.CountApprovedDeposits(ctx, evt.GetMemberID(), window)
		if err != nil {
			return 0, 0, err
		}
		sum, err := c.Store.SumDeposits(ctx, evt.GetMemberID(), window)
		if err != nil {
			return 0, 0, err
		}
		return n, sum, nil
	}

	if c.Counter == nil {
		return load(ctx)
	}

	key := depositWindowKey(activity.ID, evt.GetMemberID(), window)
	return c.Counter.IncrOrLoad(ctx, key, evt.GetOrderID(), evt.GetAmount(), windowTTL(window), load)
}

func depositWindowKey(promoID, userID int64, window types.TimeWindow) string {
	return fmt.Sprintf("promo:dep:win:%d:%d:%d:%d",
		promoID, userID, window.Start.Unix(), window.End.Unix())
}

func windowTTL(window types.TimeWindow) time.Duration {
	ttl := time.Until(window.End) + time.Hour
	if ttl < time.Hour {
		ttl = 24 * time.Hour
	}
	const maxTTL = 40 * 24 * time.Hour
	if ttl > maxTTL {
		ttl = maxTTL
	}
	return ttl
}
