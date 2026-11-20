package kafka

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
	xkafka "oa.98ent.com/p9/common/xkafka"
	"oa.98ent.com/p9/common/xkafka/payload"
)

// Producer 基于 xkafka 的 Kafka 生产者
type Producer struct {
	producer *xkafka.Producer
}

// NewProducer 创建新的生产者实例
func NewProducer(brokers []string) *Producer {
	return &Producer{
		producer: xkafka.NewProducer(brokers, nil),
	}
}

// Publish 发送事件到指定 topic
// 将 types.Event 包装为 xkafka.Payload 后发送
func (p *Producer) Publish(
	ctx context.Context,
	topic string,
	event payload.SingleMemberEvent,
) error {
	// 序列化事件为 Event 的 JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		logx.Errorf("[kafka生产者] 序列化事件失败: err=%v", err)
		return err
	}

	// 构造 xkafka.Payload：将 Event JSON 作为 Data
	kafkaPayload := xkafka.Payload{
		Name:      event.EventType, // 使用 EventType 作为 Payload 的 Name
		Data:      eventData,       // Event 的 JSON 作为 Data
		Timestamp: event.PublishedAt.UnixMilli(),
		V:         1,
	}

	// 序列化 Payload 为 JSON
	payloadData, err := json.Marshal(kafkaPayload)
	if err != nil {
		logx.Errorf("[kafka生产者] 序列化payload失败: err=%v", err)
		return err
	}

	logx.Infof("[kafka生产者] 发送事件: topic=%s, event_type=%s, operator_code=%s, member_id=%d",
		topic, event.EventType, event.OperatorCode, event.MemberID)

	// 使用事件的 OperatorCode 作为 key，保证同一个运营商的事件在同一个分区
	return p.producer.PushWithKey(ctx, topic, event.OperatorCode, string(payloadData))
}

// Close 关闭生产者
func (p *Producer) Close() error {
	// xkafka.Producer 不需要显式关闭
	logx.Info("Kafka Producer closed")
	return nil
}
