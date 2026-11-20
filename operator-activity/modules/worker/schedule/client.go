package schedule

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/temporal"

	"github.com/bytedance/gopkg/util/logger"
)

// RunDemo 运行定时任务 demo
// 这个函数演示如何创建和管理基于 Temporal Schedule 的定时任务
func RunDemo() error {
	c, err := temporal.GetClient()
	if err != nil {
		return fmt.Errorf("创建 Temporal 客户端失败: %w", err)
	}

	// 创建定时任务管理器
	manager := NewScheduleManager(c)
	ctx := context.Background()

	// 示例1: 创建每5分钟执行一次的清理过期订单任务
	cleanupTask := TaskRequest{
		TaskID:   "cleanup-expired-orders-001",
		TaskName: "cleanup_expired_orders",
		Params: map[string]string{
			"max_age_days": "30",
		},
	}
	err = manager.CreateSchedule(ctx, "cleanup-expired-orders-schedule", "*/1 * * * * *", cleanupTask)
	if err != nil {
		logger.Error("创建清理过期订单定时任务失败", "error", err)
	} else {
		logger.Info("成功创建清理过期订单定时任务")
	}

	// 示例2: 创建每天凌晨2点执行的生成每日报告任务
	dailyReportTask := TaskRequest{
		TaskID:   "generate-daily-report-001",
		TaskName: "generate_daily_report",
		Params: map[string]string{
			"report_type": "daily",
		},
	}
	err = manager.CreateSchedule(ctx, "generate-daily-report-schedule", "0 0 2 * * *", dailyReportTask)
	if err != nil {
		logger.Error("创建生成每日报告定时任务失败", "error", err)
	} else {
		logger.Info("成功创建生成每日报告定时任务")
	}

	// 示例3: 创建每小时执行一次的数据同步任务
	syncTask := TaskRequest{
		TaskID:   "sync-data-001",
		TaskName: "sync_data",
		Params: map[string]string{
			"source": "external_api",
		},
	}
	err = manager.CreateSchedule(ctx, "sync-data-schedule", "0 0 * * * *", syncTask)
	if err != nil {
		logger.Error("创建数据同步定时任务失败", "error", err)
	} else {
		logger.Info("成功创建数据同步定时任务")
	}

	// 等待一段时间，让任务执行
	logger.Info("等待定时任务执行...")
	time.Sleep(10 * time.Second)

	// 示例4: 获取定时任务计划信息
	description, err := manager.GetSchedule(ctx, "cleanup-expired-orders-schedule")
	if err != nil {
		logger.Error("获取定时任务计划信息失败", "error", err)
	} else {
		logger.Info("定时任务计划信息", "description", description)
	}

	// 示例5: 暂停定时任务
	err = manager.PauseSchedule(ctx, "sync-data-schedule", "临时暂停数据同步任务")
	if err != nil {
		logger.Error("暂停定时任务失败", "error", err)
	} else {
		logger.Info("成功暂停数据同步定时任务")
	}

	// 等待一段时间
	time.Sleep(5 * time.Second)

	// 示例6: 恢复定时任务
	err = manager.UnpauseSchedule(ctx, "sync-data-schedule")
	if err != nil {
		logger.Error("恢复定时任务失败", "error", err)
	} else {
		logger.Info("成功恢复数据同步定时任务")
	}

	logger.Info("定时任务 demo 执行完成")
	return nil
}

// CreateScheduleFromConfig 从配置创建定时任务
// 这个函数可以从配置文件或数据库中读取定时任务配置，然后创建对应的 schedule
func CreateScheduleFromConfig() error {
	time.Sleep(5 * time.Second)
	// 创建 Temporal 客户端
	c, err := temporal.GetClient()
	if err != nil {
		return fmt.Errorf("获取 Temporal 客户端失败: %w", err)
	}

	// 创建定时任务管理器
	manager := NewScheduleManager(c)
	ctx := context.Background()

	// 这里可以从配置文件或数据库中读取定时任务配置
	scheduleConfigs := []struct {
		ScheduleID     string
		CronExpression string
		Task           TaskRequest
	}{
		{
			ScheduleID:     "installment-today-performance-schedule",
			CronExpression: "*/5 * * * *", // 每5分钟执行一次
			Task: TaskRequest{
				TaskID:   "today-performance-001",
				TaskName: "today_performance",
				Params:   map[string]string{"report_type": "daily"},
			},
		},
		{
			ScheduleID:     "installment-today-overdue-schedule",
			CronExpression: "0 2 * * *", // 每天凌晨2点
			Task: TaskRequest{
				TaskID:   "today-overdue-001",
				TaskName: "today_overdue",
				Params:   map[string]string{"report_type": "daily"},
			},
		},
	}

	// 创建所有定时任务
	for _, config := range scheduleConfigs {
		err := manager.CreateSchedule(ctx, config.ScheduleID, config.CronExpression, config.Task)
		if err != nil {
			logger.Error("创建定时任务失败",
				"schedule_id", config.ScheduleID,
				"error", err)
			continue
		}
		logger.Info("成功创建定时任务", "schedule_id", config.ScheduleID)
	}

	return nil
}
