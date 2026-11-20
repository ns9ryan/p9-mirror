# Temporal Schedule 定时任务 Demo

这个目录包含了一个基于 Temporal Schedule 的定时任务 demo，演示了如何使用 Temporal 的 Schedule 功能来创建和管理定时任务。

## 功能特性

- ✅ 创建定时任务计划（支持 Cron 表达式）
- ✅ 暂停/恢复定时任务
- ✅ 删除定时任务
- ✅ 查询定时任务信息
- ✅ 多种示例任务（清理过期订单、生成每日报告、数据同步等）

## 文件说明

- `workflow.go`: 定时任务工作流定义
- `activity.go`: 定时任务活动实现（实际执行的任务逻辑）
- `manager.go`: 定时任务管理器，提供创建、删除、暂停、恢复等管理功能
- `register.go`: Worker 注册器，用于注册工作流和活动
- `demo.go`: 使用示例代码

## 使用方法

### 1. 注册 Worker

在 `biz/worker/init.go` 中添加 schedule worker 注册器：

```go
func (wm *WorkerManager) getAllWorkerRegistrars() []WorkerRegistrar {
	return []WorkerRegistrar{
		order.NewRegistrar(),
		schedule.NewRegistrar(), // 添加这一行
	}
}
```

### 2. 创建定时任务

```go
import "orderService/biz/schedule"

// 创建 Temporal 客户端
c, err := temporal.GetClient()
if err != nil {
    log.Fatal(err)
}

// 创建定时任务管理器
manager := schedule.NewScheduleManager(c)

// 创建定时任务
task := schedule.TaskRequest{
    TaskID:   "cleanup-expired-orders-001",
    TaskName: "cleanup_expired_orders",
    Params:   map[string]string{"max_age_days": "30"},
}

err = manager.CreateSchedule(
    context.Background(),
    "cleanup-expired-orders-schedule", // schedule ID
    "0 */5 * * * *",                   // Cron 表达式：每5分钟执行一次
    task,
)
```

### 3. Cron 表达式说明

Temporal 使用标准的 Cron 表达式格式：`秒 分 时 日 月 星期`

示例：
- `0 */5 * * * *` - 每5分钟执行一次
- `0 0 2 * * *` - 每天凌晨2点执行
- `0 0 * * * *` - 每小时执行一次
- `0 0 0 * * *` - 每天午夜执行
- `0 0 9 * * 1-5` - 每个工作日上午9点执行

### 4. 管理定时任务

```go
// 暂停定时任务
err = manager.PauseSchedule(ctx, "schedule-id", "暂停原因")

// 恢复定时任务
err = manager.UnpauseSchedule(ctx, "schedule-id")

// 删除定时任务
err = manager.DeleteSchedule(ctx, "schedule-id")

// 获取定时任务信息
description, err := manager.GetSchedule(ctx, "schedule-id")
```

### 5. 运行 Demo

```go
// 运行完整的 demo
err := schedule.RunDemo()

// 或从配置创建定时任务
err := schedule.CreateScheduleFromConfig()
```

## 添加自定义任务

在 `activity.go` 中添加新的任务处理函数：

```go
func (a *ScheduledTaskActivities) myCustomTask(ctx context.Context, task TaskRequest) (*TaskResponse, error) {
    // 实现你的任务逻辑
    return &TaskResponse{
        TaskID:     task.TaskID,
        TaskName:   task.TaskName,
        Success:    true,
        Message:    "任务执行成功",
        ExecutedAt: time.Now(),
    }, nil
}
```

然后在 `ExecuteTask` 方法中添加对应的 case：

```go
case "my_custom_task":
    result, err = a.myCustomTask(ctx, task)
```

## 注意事项

1. **Worker 必须运行**: 确保 schedule worker 已经启动并运行，否则定时任务无法执行
2. **Cron 表达式格式**: Temporal 使用 6 位 Cron 表达式（包含秒）
3. **时区**: 默认使用 UTC 时区，如需指定时区，可以在创建 Schedule 时配置
4. **任务幂等性**: 确保任务逻辑是幂等的，因为可能会重试执行

## 相关文档

- [Temporal Schedule 官方文档](https://docs.temporal.io/workflows#schedules)
- [Cron 表达式参考](https://en.wikipedia.org/wiki/Cron)
