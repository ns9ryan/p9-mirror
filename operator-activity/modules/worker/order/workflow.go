package order

import (
	"errors"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/constants"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// StartActivityOrderWorkflow 启动活动订单工作流
func StartActivityOrderWorkflow(ctx workflow.Context, order CreateOrderReq) (*CreateOrderReq, error) {
	logger := workflow.GetLogger(ctx)
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 2,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger.Info("活动订单工作流开始", "order_no", order.OrderNo, "workflow_id", workflowID, "payout_mode", order.PayoutMode)

	orderStatus := order.Status
	_ = workflow.SetQueryHandler(ctx, "getOrderStatus", func() (types.OrderStatus, error) {
		return orderStatus, nil
	})
	_ = workflow.SetQueryHandler(ctx, "getOrderInfo", func() (*CreateOrderReq, error) {
		order.Status = orderStatus
		return &order, nil
	})
	// 订单活动实例
	var activities *OrderActivities
	// 如果订单状态为待审核，则等待人工审核信号
	if order.Status == types.OrderStatusPendingAudit {
		userAuditSignalChan := workflow.GetSignalChannel(ctx, constants.WorkflowUserAuditSignal)
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(userAuditSignalChan, func(channel workflow.ReceiveChannel, more bool) {
			var req UserAuditSignalReq
			channel.Receive(ctx, &req)
			if req.OrderNo != order.OrderNo {
				return
			}
			if req.Result == "approved" {
				orderStatus = types.OrderStatusApproved
			} else {
				orderStatus = types.OrderStatusRejected
			}
			order.Status = orderStatus
			order.Remark = req.Reason
			logger.Info("收到用户审核信号", "order_no", order.OrderNo, "result", req.Result, "reason", req.Reason)
		})
		selector.Select(ctx)
		// 执行订单审核
		if err := workflow.ExecuteActivity(ctx, activities.AuditOrder, order).Get(ctx, nil); err != nil {
			logger.Error("订单审核失败", "error", err)
			return nil, err
		}
		// 如果订单状态为拒绝，则直接返回
		if order.Status == types.OrderStatusRejected {
			return &order, nil
		}
	}

	switch order.PayoutMode {
	case types.PayoutAuto:
		if err := waitUntil(ctx, order); err != nil {
			return nil, err
		}
		order.PayoutAt = workflow.Now(ctx)
		order.Remark = "自动派发到账"
		orderStatus = types.OrderStatusPaidOut
		order.Status = orderStatus
		if err := workflow.ExecuteActivity(ctx, activities.DisburseOrder, order).Get(ctx, nil); err != nil {
			logger.Error("自动派发失败", "error", err)
			return nil, err
		}
	case types.PayoutManualExpireVoid, types.PayoutManualExpireAuto:
		// 次日可领：先等到 PayoutAt 再开领取窗口
		if err := waitUntil(ctx, order); err != nil {
			return nil, err
		}
		// 如果到了领取时间订单状态为过审，则设置为待领取
		if order.Status == types.OrderStatusApproved {
			orderStatus = types.OrderStatusPendingClaim
			order.Status = orderStatus
		}

		// 以配置的 ClaimExpireSec 作为领取窗口时长，从「可领时刻」起算，避免审核/Activity 耗时压缩窗口
		var claimTimeout time.Duration
		if !order.ClaimExp.IsZero() && !order.PayoutAt.IsZero() {
			claimTimeout = max(order.ClaimExp.Sub(order.PayoutAt), 0)
		}
		order.ClaimExp = workflow.Now(ctx).Add(claimTimeout)

		if order.Status == types.OrderStatusPendingClaim {
			if err := workflow.ExecuteActivity(ctx, activities.PendingClaimOrder, order).Get(ctx, nil); err != nil {
				logger.Error("更新订单状态失败", "error", err)
				return nil, err
			}
		}

		userClaimSignalChan := workflow.GetSignalChannel(ctx, constants.WorkflowUserClaimSignal)
		timerCtx, cancelOrderTimer := workflow.WithCancel(ctx)
		timerFuture := workflow.NewTimer(timerCtx, claimTimeout)
		selector := workflow.NewSelector(ctx)

		var claimed bool
		selector.AddReceive(userClaimSignalChan, func(channel workflow.ReceiveChannel, more bool) {
			var orderNo string
			channel.Receive(ctx, &orderNo)
			cancelOrderTimer()
			claimed = true
			logger.Info("收到用户领取信号", "order_no", orderNo)
		})
		selector.AddFuture(timerFuture, func(f workflow.Future) {
			_ = f.Get(ctx, nil)
			logger.Info("活动订单用户超时未领取", "order_no", order.OrderNo, "timeout_duration", claimTimeout)
		})
		selector.Select(ctx)

		if claimed {
			order.PayoutAt = workflow.Now(ctx)
			order.Remark = "用户手动领取"
			orderStatus = types.OrderStatusPaidOut
			order.Status = orderStatus
			if err := workflow.ExecuteActivity(ctx, activities.ClaimOrder, order).Get(ctx, nil); err != nil {
				logger.Error("用户领取失败", "error", err)
				return nil, err
			}
			break
		}

		switch order.PayoutMode {
		case types.PayoutManualExpireVoid:
			order.Status = types.OrderStatusExpired
			order.Remark = "活动订单用户超时未领取过期作废"
			order.CancelledAt = ToPtr(workflow.Now(ctx))
			orderStatus = types.OrderStatusExpired
			if err := workflow.ExecuteActivity(ctx, activities.TimeoutCancelOrder, order, claimTimeout).Get(ctx, nil); err != nil {
				logger.Error("超时作废失败", "error", err)
				return nil, err
			}
		case types.PayoutManualExpireAuto:
			order.Status = types.OrderStatusPaidOut
			order.Remark = "活动订单用户超时自动领取"
			order.PayoutAt = workflow.Now(ctx)
			orderStatus = types.OrderStatusPaidOut
			if err := workflow.ExecuteActivity(ctx, activities.TimeoutAutoClaimOrder, order, claimTimeout).Get(ctx, nil); err != nil {
				logger.Error("超时自动领取失败", "error", err)
				return nil, err
			}
		}
	default:
		return nil, errors.New("payout_mode not supported")
	}

	logger.Info("活动订单工作流结束", "order_id", order.ID, "status", orderStatus)
	return &order, nil
}

func waitUntil(ctx workflow.Context, order CreateOrderReq) error {
	// 如果订单状态为待领取，则直接返回
	// if order.Status == types.OrderStatusPendingClaim {
	// 	return nil
	// }
	// 如果订单支付时间或创建时间为空，则直接返回
	if order.PayoutAt.IsZero() || order.CreatedAt.IsZero() {
		return nil
	}
	// d := order.PayoutAt.Sub(order.CreatedAt)
	d := order.PayoutAt.Sub(workflow.Now(ctx))
	if d <= 0 {
		return nil
	}
	return workflow.NewTimer(ctx, d).Get(ctx, nil)
}

func getCreateOrderWorkflowID(opCode string, orderNo string) string {
	return fmt.Sprintf("%s-%s", opCode, orderNo)
}

// 将值转换为指针
func ToPtr[T any](v T) *T {
	return new(v)
}
