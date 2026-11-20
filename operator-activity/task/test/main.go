package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/xkafka"
	"oa.98ent.com/p9/operator-activity/task/constants"
)

// Event 测试事件结构（独立定义，不依赖项目）
type Event struct {
	EventType    string    `json:"event_type"`
	OperatorCode string    `json:"operator_code"`
	MemberID     int64     `json:"member_id"`
	OccurredAt   time.Time `json:"occurred_at"`
	PublishedAt  time.Time `json:"published_at"`
	TraceID      string    `json:"trace_id"`
}

var (
	kafkaAddr    = flag.String("kafka", "localhost:9094", "Kafka broker address")
	operatorCode = flag.String("op", "OP_61AAD9", "Operator code for testing")
	memberID     = flag.Int64("member", 100001, "Test member ID")
	eventCount   = flag.Int("count", 3, "Number of events to send per topic")
	intervalMs   = flag.Int("interval", 5000, "Interval in milliseconds between sending events")
	topic        = flag.String("topic", "", "Specific topic to send (empty = all topics)")
)

// defaultTopics 默认测试主题
var defaultTopics = constants.TopicToEventType

func main() {
	flag.Parse()

	logx.Infof("🚀 独立 Kafka 测试数据生成器启动")
	logx.Infof("📊 配置: kafka=%s, operator=%s, member_id=%d", *kafkaAddr, *operatorCode, *memberID)

	ctx := context.Background()

	// 确定要发送的主题
	topicsToSend := defaultTopics
	if *topic != "" {
		topicsToSend = map[string]string{*topic: *topic}
	}

	logx.Infof("✅ 已连接到 Kafka: %s", *kafkaAddr)
	fmt.Printf("✅ 已连接到 Kafka broker\n\n")
	OccurredAt := time.Now().Add(-2 * 24 * time.Hour)
	// 发送测试数据
	totalSent := 0
	for kafkaTopic, eventType := range topicsToSend {
		// 为每个 topic 创建 writer
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{*kafkaAddr},
			Topic:    kafkaTopic,
			Balancer: &kafka.LeastBytes{},
		})
		defer writer.Close()

		if eventType == constants.EventTypeMemberRegistered || eventType == constants.EventTypeDepositSucceeded {
			batchMessages := []kafka.Message{}
			batchSize := 10

			for i := 1; i <= *eventCount; i++ {
				event := Event{
					EventType:    eventType,
					OperatorCode: *operatorCode,
					MemberID:     *memberID + int64(i),
					OccurredAt:   OccurredAt,
					PublishedAt:  time.Now(),
					TraceID:      uuid.New().String(),
				}

				// 序列化事件
				eventBytes, err := json.Marshal(event)
				if err != nil {
					logx.Errorf("❌ 序列化事件失败: %v", err)
					continue
				}

				kafkaPayload := xkafka.Payload{
					Name:      event.EventType, // 使用 EventType 作为 Payload 的 Name
					Data:      eventBytes,      // Event 的 JSON 作为 Data
					Timestamp: event.PublishedAt.UnixMilli(),
					V:         1,
				}

				// 序列化 Payload 为 JSON
				payloadData, err := json.Marshal(kafkaPayload)
				if err != nil {
					logx.Errorf("[kafka生产者] 序列化payload失败: err=%v", err)
					continue
				}

				// 添加到批处理队列
				batchMessages = append(batchMessages, kafka.Message{
					Key:   []byte(fmt.Sprintf("%s", event.TraceID)),
					Value: payloadData,
				})

				// 达到批大小或最后一条时，批量发送
				if len(batchMessages) >= batchSize || i == *eventCount {
					err := writer.WriteMessages(ctx, batchMessages...)
					if err != nil {
						logx.Errorf("❌ 批量发送到 Kafka 失败 (topic=%s): %v", kafkaTopic, err)
						fmt.Printf("❌ 批量发送失败: topic=%s, error=%v\n", kafkaTopic, err)
					} else {
						logx.Infof("✅ 批量发送成功: topic=%s, 批次大小=%d, 总数=%d/%d",
							kafkaTopic, len(batchMessages), i, *eventCount)
						fmt.Printf("✅ 批量发送成功 [%d/%d]: topic=%s, 批次大小=%d\n", i, *eventCount, kafkaTopic, len(batchMessages))
						totalSent += len(batchMessages)
					}

					// 清空批处理队列
					batchMessages = []kafka.Message{}

					// 按指定间隔发送
					// if i < *eventCount {
					// 	time.Sleep(time.Duration(*intervalMs) * time.Millisecond)
					// }
				}
			}
			time.Sleep(time.Duration(*intervalMs) * time.Millisecond)
		}
	}

	logx.Infof("🎉 完成! 总共发送 %d 条消息", totalSent)
	fmt.Printf("🎉 完成! 总共发送 %d 条消息\n", totalSent)
}
