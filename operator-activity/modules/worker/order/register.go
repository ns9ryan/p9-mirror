package order

import (
	"oa.98ent.com/p9/operator-activity/constants"
	"oa.98ent.com/p9/operator-activity/ent"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Registrar 创建订单 worker 注册器
type Registrar struct{}

// NewRegistrar 创建创建订单 worker 注册器
func NewRegistrar() *Registrar {
	return &Registrar{}
}

// Register 注册结算 worker
func (r *Registrar) Register(c client.Client, entClient *ent.Client) worker.Worker {
	w := worker.New(c, constants.WorkflowQueueName, worker.Options{
		MaxConcurrentWorkflowTaskPollers: 8, // 默认 2
		MaxConcurrentActivityTaskPollers: 8, // 默认 2
	})
	w.RegisterWorkflow(StartActivityOrderWorkflow)
	// 注册活动
	a := &OrderActivities{client: entClient}
	w.RegisterActivity(a.TimeoutCancelOrder)
	w.RegisterActivity(a.PendingClaimOrder)
	w.RegisterActivity(a.TimeoutAutoClaimOrder)
	w.RegisterActivity(a.ClaimOrder)
	w.RegisterActivity(a.AuditOrder)
	w.RegisterActivity(a.DisburseOrder)
	return w
}

// Name 返回 worker 名称
func (r *Registrar) Name() string {
	return "order"
}
