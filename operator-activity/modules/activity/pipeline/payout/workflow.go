package payout

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// WorkflowStarter 启动订单领取/派发 Temporal 工作流（由 worker 实现，nil 时跳过）
type WorkflowStarter interface {
	StartOrderLifecycle(ctx context.Context, order *types.ActivityOrder) error
}
