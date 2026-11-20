package schedule

import (
	"oa.98ent.com/p9/operator-activity/ent"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Registrar 定时任务 worker 注册器
type Registrar struct{}

// NewRegistrar 创建定时任务 worker 注册器
func NewRegistrar() *Registrar {
	return &Registrar{}
}

// Register 注册定时任务 worker
func (r *Registrar) Register(c client.Client, entClient *ent.Client) worker.Worker {
	w := worker.New(c, ScheduleTaskQueue, worker.Options{})

	// 注册工作流
	w.RegisterWorkflow(ScheduledTaskWorkflow)

	// 注册活动
	a := &ScheduledTaskActivities{
		client: entClient,
	}
	w.RegisterActivity(a.ExecuteTask)

	return w
}

// Name 返回 worker 名称
func (r *Registrar) Name() string {
	return "schedule"
}
