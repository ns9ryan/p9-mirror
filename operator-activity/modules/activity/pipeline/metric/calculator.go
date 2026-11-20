package metric

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// Calculator 触发指标计算策略
type Calculator interface {
	CalcMetric(ctx context.Context, activity *types.Activity, evt types.ActivityEvent) (*types.TriggerMetric, error)
}

// MetricStore 指标查询存储
type MetricStore interface {
	CountApprovedDeposits(ctx context.Context, memberID int64, since types.TimeWindow) (int, error)
	SumDeposits(ctx context.Context, memberID int64, window types.TimeWindow) (int64, error)
	SumBets(ctx context.Context, memberID int64, window types.TimeWindow) (int64, error)
	CountBets(ctx context.Context, memberID int64, window types.TimeWindow) (int64, error)
	CountValidReferrals(ctx context.Context, evt types.ActivityEvent, activity *types.Activity, window types.TimeWindow) (int, int, int, error)
}
