package settlement

import (
	"oa.98ent.com/p9/operator-activity/ent"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	// QueueName 结算任务队列名称
	QueueName = "settlement_queue"
	// WorkflowName 结算工作流名称
	WorkflowName = "JobSettlementWorkflow"
)

// Registrar 结算 worker 注册器
type Registrar struct{}

// NewRegistrar 创建结算 worker 注册器
func NewRegistrar() *Registrar {
	return &Registrar{}
}

// Register 注册结算 worker
func (r *Registrar) Register(c client.Client, _ *ent.Client) worker.Worker {
	w := worker.New(c, QueueName, worker.Options{})
	w.RegisterWorkflow(JobSettlementWorkflow)
	w.RegisterActivity(DoSettlementActivity)
	return w
}

// Name 返回 worker 名称
func (r *Registrar) Name() string {
	return "settlement"
}
