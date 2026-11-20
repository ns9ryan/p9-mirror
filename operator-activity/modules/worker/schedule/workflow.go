package schedule

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ScheduledTaskWorkflow 定时任务工作流
// 这个工作流会被 schedule 定期触发执行
func ScheduledTaskWorkflow(ctx workflow.Context, task TaskRequest) (*TaskResponse, error) {
	logger := workflow.GetLogger(ctx)

	// 记录工作流开始时间
	startTime := workflow.Now(ctx)
	logger.Info("定时任务工作流开始执行",
		"task_id", task.TaskID,
		"task_name", task.TaskName,
		"start_time", startTime)

	// 设置活动选项
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 6,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	// 设置时区为Asia/Shanghai
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logger.Warn("加载时区失败，使用 UTC", "error", err)
		location = time.UTC
	}
	ctx = workflow.WithValue(ctx, "location", location)

	// 执行定时任务活动
	var activities *ScheduledTaskActivities
	var result TaskResponse

	err = workflow.ExecuteActivity(ctx, activities.ExecuteTask, task).Get(ctx, &result)
	if err != nil {
		logger.Error("执行定时任务活动失败",
			"task_id", task.TaskID,
			"error", err)
		return nil, fmt.Errorf("执行定时任务失败: %w", err)
	}

	// 记录工作流结束时间
	endTime := workflow.Now(ctx)
	duration := endTime.Sub(startTime)
	logger.Info("定时任务工作流执行完成",
		"task_id", task.TaskID,
		"task_name", task.TaskName,
		"end_time", endTime,
		"duration", duration,
		"result", result)

	return &result, nil
}

// TaskRequest 任务请求参数
type TaskRequest struct {
	TaskID   string            `json:"task_id"`
	TaskName string            `json:"task_name"`
	Params   map[string]string `json:"params"`
}

// TaskResponse 任务响应结果
type TaskResponse struct {
	TaskID     string    `json:"task_id"`
	TaskName   string    `json:"task_name"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	ExecutedAt time.Time `json:"executed_at"`
}
