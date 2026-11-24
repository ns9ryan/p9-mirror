package kafka

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/client/taskservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/agentdepositorderservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/member"
)

// EventHandler Kafka事件处理器
type EventHandler struct {
	producer             *Producer
	memberRpc            member.Member
	agentDepositorderRpc agentdepositorderservice.AgentDepositOrderService
	taskRpc              taskservice.TaskService
	logx.Logger
}

// NewEventHandler 创建事件处理器
func NewEventHandler(producer *Producer, memberRpc member.Member, agentDepositorderRpc agentdepositorderservice.AgentDepositOrderService, taskRpc taskservice.TaskService) *EventHandler {
	return &EventHandler{
		producer:             producer,
		memberRpc:            memberRpc,
		agentDepositorderRpc: agentDepositorderRpc,
		taskRpc:              taskRpc,
		Logger:               logx.WithContext(context.Background()),
	}
}
