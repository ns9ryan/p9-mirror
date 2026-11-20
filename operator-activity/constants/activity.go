package constants

const (
	ActivityOrderPrefix = "PAO" // 活动订单前缀
)

// 活动订单工作流常量
const (
	// 订单工作流任务队列名称
	WorkflowQueueName = "order_queue"
	// 用户领取信号
	WorkflowUserClaimSignal = "userClaim"
	// 用户审核信号
	WorkflowUserAuditSignal = "userAudit"
)
