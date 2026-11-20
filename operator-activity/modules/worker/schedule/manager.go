package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"go.temporal.io/sdk/client"
)

const (
	// ScheduledTaskWorkflowName 定时任务工作流名称
	ScheduledTaskWorkflowName = "ScheduledTaskWorkflow"
	// ScheduleTaskQueue 定时任务队列名称
	ScheduleTaskQueue = "schedule_task_queue"
)

// ScheduleManager 定时任务管理器
type ScheduleManager struct {
	client client.Client
}

// NewScheduleManager 创建定时任务管理器
func NewScheduleManager(c client.Client) *ScheduleManager {
	return &ScheduleManager{
		client: c,
	}
}

// CreateSchedule 创建定时任务计划
// scheduleID: 计划唯一标识
// cronExpression: Cron 表达式，例如 "0 */5 * * * *" 表示每5分钟执行一次
// task: 要执行的任务参数
func (sm *ScheduleManager) CreateSchedule(ctx context.Context, scheduleID, cronExpression string, task TaskRequest) error {
	logger.Info("创建定时任务计划",
		"schedule_id", scheduleID,
		"cron_expression", cronExpression,
		"task_name", task.TaskName)

	// 设置时区为 Asia/Shanghai (UTC+8)
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logger.Warn("加载时区失败，使用 UTC", "error", err)
		location = time.UTC
	}
	// 创建 schedule 规范
	scheduleSpec := client.ScheduleSpec{
		CronExpressions: []string{cronExpression},
		TimeZoneName:    location.String(), // 设置时区
	}

	// 创建 schedule 动作（执行工作流）
	scheduleAction := &client.ScheduleWorkflowAction{
		ID:        fmt.Sprintf("%s-workflow", scheduleID),
		Workflow:  ScheduledTaskWorkflowName,
		TaskQueue: ScheduleTaskQueue,
		Args:      []interface{}{task},
	}

	// 创建 schedule 选项
	scheduleOptions := client.ScheduleOptions{
		ID:     scheduleID,
		Spec:   scheduleSpec,
		Action: scheduleAction,
	}

	// 创建 schedule
	_, err = sm.client.ScheduleClient().Create(ctx, scheduleOptions)
	if err != nil {
		logger.Error("创建定时任务计划失败",
			"schedule_id", scheduleID,
			"error", err)
		return fmt.Errorf("创建定时任务计划失败: %w", err)
	}

	logger.Info("定时任务计划创建成功", "schedule_id", scheduleID)
	return nil
}

// DeleteSchedule 删除定时任务计划
func (sm *ScheduleManager) DeleteSchedule(ctx context.Context, scheduleID string) error {
	logger.Info("删除定时任务计划", "schedule_id", scheduleID)

	scheduleHandle := sm.client.ScheduleClient().GetHandle(ctx, scheduleID)
	err := scheduleHandle.Delete(ctx)
	if err != nil {
		logger.Error("删除定时任务计划失败",
			"schedule_id", scheduleID,
			"error", err)
		return fmt.Errorf("删除定时任务计划失败: %w", err)
	}

	logger.Info("定时任务计划删除成功", "schedule_id", scheduleID)
	return nil
}

// PauseSchedule 暂停定时任务计划
func (sm *ScheduleManager) PauseSchedule(ctx context.Context, scheduleID string, note string) error {
	logger.Info("暂停定时任务计划",
		"schedule_id", scheduleID,
		"note", note)

	scheduleHandle := sm.client.ScheduleClient().GetHandle(ctx, scheduleID)
	err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(in client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			desc := in.Description
			update := &client.ScheduleUpdate{}
			// 复制现有配置
			schedule := desc.Schedule
			update.Schedule = &schedule
			// 设置为暂停状态
			update.Schedule.State.Paused = true
			if note != "" {
				update.Schedule.State.Note = note
			}
			return update, nil
		},
	})
	if err != nil {
		logger.Error("暂停定时任务计划失败",
			"schedule_id", scheduleID,
			"error", err)
		return fmt.Errorf("暂停定时任务计划失败: %w", err)
	}

	logger.Info("定时任务计划暂停成功", "schedule_id", scheduleID)
	return nil
}

// UnpauseSchedule 恢复定时任务计划
func (sm *ScheduleManager) UnpauseSchedule(ctx context.Context, scheduleID string) error {
	logger.Info("恢复定时任务计划", "schedule_id", scheduleID)

	scheduleHandle := sm.client.ScheduleClient().GetHandle(ctx, scheduleID)
	err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(in client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			desc := in.Description
			update := &client.ScheduleUpdate{}
			// 复制现有配置
			schedule := desc.Schedule
			update.Schedule = &schedule
			// 设置为非暂停状态
			update.Schedule.State.Paused = false
			return update, nil
		},
	})
	if err != nil {
		logger.Error("恢复定时任务计划失败",
			"schedule_id", scheduleID,
			"error", err)
		return fmt.Errorf("恢复定时任务计划失败: %w", err)
	}

	logger.Info("定时任务计划恢复成功", "schedule_id", scheduleID)
	return nil
}

// GetSchedule 获取定时任务计划信息
func (sm *ScheduleManager) GetSchedule(ctx context.Context, scheduleID string) (*client.ScheduleDescription, error) {
	scheduleHandle := sm.client.ScheduleClient().GetHandle(ctx, scheduleID)
	description, err := scheduleHandle.Describe(ctx)
	if err != nil {
		logger.Error("获取定时任务计划信息失败",
			"schedule_id", scheduleID,
			"error", err)
		return nil, fmt.Errorf("获取定时任务计划信息失败: %w", err)
	}

	return description, nil
}

// ListSchedules 列出所有定时任务计划（需要 Temporal 1.23+ 版本支持）
// 注意：Temporal SDK 可能不直接支持列出所有 schedules，这里提供一个示例
func (sm *ScheduleManager) ListSchedules(ctx context.Context) ([]string, error) {
	// 注意：Temporal SDK 可能不直接支持列出所有 schedules
	// 实际使用时，建议在数据库中维护 schedule 列表
	logger.Warn("ListSchedules 功能需要 Temporal 1.23+ 版本支持，或使用数据库维护 schedule 列表")
	return nil, fmt.Errorf("暂不支持列出所有 schedules，请使用数据库维护 schedule 列表")
}
