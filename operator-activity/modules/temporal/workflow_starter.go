package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// WorkflowOptions workflow 启动选项
type WorkflowOptions struct {
	// WorkflowID 工作流ID，如果不提供则自动生成
	WorkflowID string
	// TaskQueue 任务队列名称
	TaskQueue string
	// WorkflowExecutionTimeout 工作流执行超时时间
	WorkflowExecutionTimeout time.Duration
	// WorkflowRunTimeout 工作流运行超时时间
	WorkflowRunTimeout time.Duration
	// WorkflowTaskTimeout 工作流任务超时时间
	WorkflowTaskTimeout time.Duration
	// RetryPolicy 重试策略
	RetryPolicy *temporal.RetryPolicy
	// CronSchedule 定时任务
	CronSchedule string
	// SearchAttributes 搜索属性
	SearchAttributes map[string]interface{}
}

// StartWorkflow 启动 workflow 的通用方法
func StartWorkflow(
	ctx context.Context,
	workflowName interface{},
	workflowOptions WorkflowOptions,
	args ...interface{},
) (client.WorkflowRun, error) {
	c, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("获取 Temporal 客户端失败: %w", err)
	}

	opts := client.StartWorkflowOptions{
		TaskQueue: workflowOptions.TaskQueue,
	}

	// 设置 WorkflowID
	if workflowOptions.WorkflowID != "" {
		opts.ID = workflowOptions.WorkflowID
	}

	// 设置超时时间
	if workflowOptions.WorkflowExecutionTimeout > 0 {
		opts.WorkflowExecutionTimeout = workflowOptions.WorkflowExecutionTimeout
	}
	if workflowOptions.WorkflowRunTimeout > 0 {
		opts.WorkflowRunTimeout = workflowOptions.WorkflowRunTimeout
	}
	if workflowOptions.WorkflowTaskTimeout > 0 {
		opts.WorkflowTaskTimeout = workflowOptions.WorkflowTaskTimeout
	}

	// 设置定时任务
	if workflowOptions.CronSchedule != "" {
		opts.CronSchedule = workflowOptions.CronSchedule
	}

	// 设置重试策略
	if workflowOptions.RetryPolicy != nil {
		opts.RetryPolicy = workflowOptions.RetryPolicy
	}

	// 设置搜索属性
	if len(workflowOptions.SearchAttributes) > 0 {
		opts.SearchAttributes = make(map[string]interface{})
		for k, v := range workflowOptions.SearchAttributes {
			opts.SearchAttributes[k] = v
		}
	}
	// 启动 workflow
	workflowRun, err := c.ExecuteWorkflow(ctx, opts, workflowName, args...)
	if err != nil {
		return nil, fmt.Errorf("启动 workflow 失败: %w", err)
	}

	return workflowRun, nil
}

// GetWorkflowResult 获取 workflow 执行结果
func GetWorkflowResult(ctx context.Context, workflowRun client.WorkflowRun, result interface{}) error {
	if err := workflowRun.Get(ctx, result); err != nil {
		return fmt.Errorf("获取 workflow 结果失败: %w", err)
	}
	return nil
}

// SignalWorkflow 向 workflow 发送信号
func SignalWorkflow(ctx context.Context, workflowID, signalName string, signalArg interface{}) error {
	c, err := GetClient()
	if err != nil {
		return fmt.Errorf("获取 Temporal 客户端失败: %w", err)
	}

	err = c.SignalWorkflow(ctx, workflowID, "", signalName, signalArg)
	if err != nil {
		return fmt.Errorf("发送 workflow 信号失败: %w", err)
	}

	return nil
}

// CancelWorkflow 取消 workflow
func CancelWorkflow(ctx context.Context, workflowID string) error {
	c, err := GetClient()
	if err != nil {
		return fmt.Errorf("获取 Temporal 客户端失败: %w", err)
	}

	err = c.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		return fmt.Errorf("取消 workflow 失败: %w", err)
	}

	return nil
}

// TerminateWorkflow 终止 workflow
func TerminateWorkflow(ctx context.Context, workflowID, reason string) error {
	c, err := GetClient()
	if err != nil {
		return fmt.Errorf("获取 Temporal 客户端失败: %w", err)
	}

	err = c.TerminateWorkflow(ctx, workflowID, "", reason)
	if err != nil {
		return fmt.Errorf("终止 workflow 失败: %w", err)
	}

	return nil
}
