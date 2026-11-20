package kafka

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// EventHandler Kafka事件处理器
type EventHandler struct {
	producer *Producer
	logx.Logger
}

// NewEventHandler 创建事件处理器
func NewEventHandler(producer *Producer) *EventHandler {
	return &EventHandler{
		producer: producer,
		Logger:   logx.WithContext(context.Background()),
	}
}
