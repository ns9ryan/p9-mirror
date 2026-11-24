package kafka

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/dao/member"
)

// EventHandler Kafka事件处理器
type EventHandler struct {
	producer  *Producer
	memberDao *member.Service
	logx.Logger
}

// NewEventHandler 创建事件处理器
func NewEventHandler(producer *Producer, memberDao *member.Service) *EventHandler {
	return &EventHandler{
		producer:  producer,
		memberDao: memberDao,
		Logger:    logx.WithContext(context.Background()),
	}
}
