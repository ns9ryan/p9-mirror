package settlement

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/temporal"

	temporalSDK "go.temporal.io/sdk/temporal"

	"go.temporal.io/sdk/client"
)

// StartSettlementWorkflow 启动结算工作流
// 这是业务模块特定的封装，提供更友好的接口
func StartSettlementWorkflow(ctx context.Context, param SettlementWorkflowParam) (client.WorkflowRun, error) {
	// 生成 WorkflowID（使用收款单ID确保唯一性）
	workflowID := "settlement_" + param.ReceiptID
	// 配置工作流选项
	options := temporal.WorkflowOptions{
		WorkflowID:               workflowID,
		TaskQueue:                QueueName,
		WorkflowExecutionTimeout: time.Hour * 24,   // 24小时执行超时
		WorkflowRunTimeout:       time.Hour * 24,   // 24小时运行超时
		WorkflowTaskTimeout:      time.Minute * 10, // 10分钟任务超时
		CronSchedule:             "",
		RetryPolicy: &temporalSDK.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 10,
			MaximumAttempts:    3,
		},
		// 可以添加搜索属性，方便后续查询
		SearchAttributes: map[string]interface{}{
			"ReceiptID":  param.ReceiptID,
			"MerchantID": param.MerchantID,
		},
	}

	// 启动工作流
	workflowRun, err := temporal.StartWorkflow(ctx, WorkflowName, options, param)
	if err != nil {
		return nil, err
	}

	return workflowRun, nil
}

// GetSettlementWorkflowResult 获取结算工作流结果
func GetSettlementWorkflowResult(ctx context.Context, workflowRun client.WorkflowRun) (*SettlementWorkflowResult, error) {
	var result SettlementWorkflowResult
	if err := workflowRun.Get(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSettlementWorkflowResultByID 根据 WorkflowID 获取结算工作流结果
func GetSettlementWorkflowResultByID(ctx context.Context, receiptID string) (*SettlementWorkflowResult, error) {
	c, err := temporal.GetClient()
	if err != nil {
		return nil, err
	}

	workflowID := "settlement_" + receiptID
	workflowRun := c.GetWorkflow(ctx, workflowID, "")

	var result SettlementWorkflowResult
	if err := workflowRun.Get(ctx, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
