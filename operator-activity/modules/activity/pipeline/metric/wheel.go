package metric

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// WheelCalculator 幸运转盘指标：按当前事件维度取累计投注或累计充值
// （活动可配置仅投注、仅充值，或两者皆可；由上层 matchTrigger 过滤事件）
type WheelCalculator struct {
	Store        MetricStore
	Cycle        *pipeline.Manager
	Counter      DepositWindowCounter // 可选，充值触发时复用
	DepositEvent types.EventType      // 插件定义的充值事件类型
}

func (c *WheelCalculator) CalcMetric(ctx context.Context, activity *types.Activity, evt types.ActivityEvent) (*types.TriggerMetric, error) {
	_, window, _ := c.Cycle.CurrentRound(activity, evt.GetEventTime())
	m := &types.TriggerMetric{}

	if evt.GetEventType() == c.DepositEvent {
		m.FromDeposit = true
		return c.calcDeposit(ctx, activity, evt, window, m)
	}
	return c.calcBet(ctx, evt, window, m)
}

func (c *WheelCalculator) calcDeposit(ctx context.Context, activity *types.Activity, evt types.ActivityEvent, window types.TimeWindow, m *types.TriggerMetric) (*types.TriggerMetric, error) {
	m.DepositAmount = evt.GetAmount()
	if c.Counter != nil && evt.GetOrderID() > 0 {
		key := depositWindowKey(activity.ID, evt.GetMemberID(), window)
		count, cum, err := c.Counter.IncrOrLoad(ctx, key, evt.GetOrderID(), evt.GetAmount(), windowTTL(window), func(ctx context.Context) (int, int64, error) {
			n, err := c.Store.CountApprovedDeposits(ctx, evt.GetMemberID(), window)
			if err != nil {
				return 0, 0, err
			}
			sum, err := c.Store.SumDeposits(ctx, evt.GetMemberID(), window)
			return n, sum, err
		})
		if err != nil {
			return nil, err
		}
		m.DepositCount = count
		m.CumulativeDeposit = cum
		return m, nil
	}
	n, err := c.Store.CountApprovedDeposits(ctx, evt.GetMemberID(), window)
	if err != nil {
		return nil, err
	}
	sum, err := c.Store.SumDeposits(ctx, evt.GetMemberID(), window)
	if err != nil {
		return nil, err
	}
	m.DepositCount = n
	m.CumulativeDeposit = sum
	return m, nil
}

func (c *WheelCalculator) calcBet(ctx context.Context, evt types.ActivityEvent, window types.TimeWindow, m *types.TriggerMetric) (*types.TriggerMetric, error) {
	m.BetAmount = evt.GetAmount()
	sum, err := c.Store.SumBets(ctx, evt.GetMemberID(), window)
	if err != nil {
		return nil, err
	}
	m.CumulativeBet = sum
	return m, nil
}
