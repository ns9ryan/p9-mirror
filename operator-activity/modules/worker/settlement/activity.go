package settlement

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
)

// SettlementActivityResult 活动返回结果
type SettlementActivityResult struct {
	SettlementTime time.Time
	Status         string
}

// DoSettlementActivity 执行实际结算逻辑
func DoSettlementActivity(ctx context.Context, param SettlementWorkflowParam) (*SettlementActivityResult, error) {
	logger := activity.GetLogger(ctx)
	//logger.Info("开始执行结算", "ReceiptID", param.ReceiptID)

	// 获取活动的执行信息
	//info := activity.GetInfo(ctx)

	// 模拟结算处理（实际应调用支付核心服务）
	settlementTime := time.Now()

	// 步骤1: 查询收款单状态（幂等检查）
	var currentStatus string
	//err := db.QueryRowContext(ctx,
	//	"SELECT status FROM receipt WHERE receipt_id = ? FOR UPDATE",
	//	param.ReceiptID,
	//).Scan(&currentStatus)

	//if err == sql.ErrNoRows {
	//	return nil, fmt.Errorf("收款单不存在: %s", param.ReceiptID)
	//}
	//if err != nil {
	//	return nil, err
	//}

	// 如果已结算，直接返回成功（幂等）
	if currentStatus == "SETTLED" {
		//logger.Info("收款单已结？算，跳过处理", "ReceiptID", param.ReceiptID)
		return &SettlementActivityResult{
			SettlementTime: settlementTime,
			Status:         "SETTLED",
		}, nil
	}

	// 步骤2: 执行资金划转（心跳上报进度）
	for i := 1; i <= 3; i++ {
		// 每30秒上报一次心跳，携带进度信息
		activity.RecordHeartbeat(ctx, fmt.Sprintf("settlement_progress_%d", i*33))
		time.Sleep(time.Second * 30) // 模拟耗时操作
	}

	// 步骤3: 更新收款单状态为已结算
	//_, err = db.ExecContext(ctx,
	//	`UPDATE receipt SET
	//		status = ?,
	//		settlement_time = ?,
	//		workflow_id = ?,
	//		settlement_amount = ?
	//	WHERE receipt_id = ? AND status != ?`,
	//	"SETTLED",
	//	settlementTime,
	//	info.WorkflowExecution.ID,
	//	param.Amount,
	//	param.ReceiptID,
	//	"SETTLED",
	//)
	//if err != nil {
	//	return nil, fmt.Errorf("更新收款单状态失败: %w", err)
	//}

	// 步骤4: 发送结算通知（可异步处理）
	go sendSettlementNotification(param.ReceiptID, param.MerchantID)

	logger.Info("结算活动完成", "ReceiptID", param.ReceiptID)

	return &SettlementActivityResult{
		SettlementTime: settlementTime,
		Status:         "SETTLED",
	}, nil
}

// 发送结算通知（模拟）
func sendSettlementNotification(receiptID, merchantID string) {
	// 调用消息中心或通知服务
	time.Sleep(time.Second * 2)
}
