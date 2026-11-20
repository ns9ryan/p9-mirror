package settlement

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SettlementWorkflowResult 工作流返回结果
type SettlementWorkflowResult struct {
	ReceiptID      string
	SettlementTime time.Time
	Status         string
}

// JobSettlementWorkflow 收款单T+1结算工作流
func JobSettlementWorkflow(ctx workflow.Context, param SettlementWorkflowParam) (*SettlementWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("收款单结算工作流启动", "ReceiptID", param.ReceiptID)

	// 1. 计算T+1日结算时间（下一个工作日的9:00）
	t1SettlementTime := calculateNextWorkday(param.ReceiptDate, 9, 0)

	// 2. 计算需要等待的时间
	now := workflow.Now(ctx)
	waitDuration := t1SettlementTime.Sub(now)

	if waitDuration > 0 {
		logger.Info("等待到T+1结算时间",
			"等待时长", waitDuration.String(),
			"目标时间", t1SettlementTime)

		// 使用workflow.Sleep等待，支持工作流恢复
		_ = workflow.Sleep(ctx, waitDuration)
	}

	// 3. 执行结算活动（配置重试策略）
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 10,
			MaximumAttempts:    5,
			NonRetryableErrorTypes: []string{
				"InvalidReceiptError",
				"InsufficientBalanceError",
			},
		},
		HeartbeatTimeout: time.Minute * 5, // 5分钟心跳超时
	})

	var settlementResult SettlementActivityResult
	err := workflow.ExecuteActivity(activityCtx, DoSettlementActivity, param).Get(ctx, &settlementResult)
	if err != nil {
		logger.Error("结算活动执行失败", "error", err)
		return nil, fmt.Errorf("结算失败: %w", err)
	}

	logger.Info("结算完成", "ReceiptID", param.ReceiptID)

	return &SettlementWorkflowResult{
		ReceiptID:      param.ReceiptID,
		SettlementTime: settlementResult.SettlementTime,
		Status:         settlementResult.Status,
	}, nil
}

// 计算下一个工作日时间（跳过周末）
func calculateNextWorkday(baseTime time.Time, hour, min int) time.Time {
	nextDay := baseTime.AddDate(0, 0, 1) // T+1日
	// 如果是周末，跳到周一
	//if nextDay.Weekday() == time.Saturday {
	//	nextDay = nextDay.AddDate(0, 0, 2)
	//} else if nextDay.Weekday() == time.Sunday {
	//	nextDay = nextDay.AddDate(0, 0, 1)
	//}
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(),
		hour, min, 0, 0, nextDay.Location())
}
