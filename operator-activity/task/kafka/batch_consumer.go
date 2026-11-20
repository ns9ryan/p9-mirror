package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	xkafka "oa.98ent.com/p9/common/xkafka"
	"oa.98ent.com/p9/operator-activity/task/constants"
	taskqueue "oa.98ent.com/p9/operator-activity/task/queue"
	"oa.98ent.com/p9/operator-activity/task/types"
	"oa.98ent.com/p9/operator-activity/task/utils"
)

// KafkaBatchConsumer Kafka 批次消费者
// 简化设计：每条消息都立即生成一个批次并写入 Redis
type KafkaBatchConsumer struct {
	redisQueue *taskqueue.RedisQueue
	topic      string
	aggregator *taskqueue.BatchAggregator
	logger     logx.Logger
}

// NewKafkaBatchConsumer 创建新的批次消费者
func NewKafkaBatchConsumer(ctx context.Context, redisQueue *taskqueue.RedisQueue, topic string) *KafkaBatchConsumer {
	logger := logx.WithContext(ctx)
	return &KafkaBatchConsumer{
		redisQueue: redisQueue,
		topic:      topic,
		aggregator: taskqueue.NewBatchAggregator(logger),
		logger:     logger,
	}
}

// Handle 处理单条 Kafka 消息
// 简化设计：每条消息都立即生成一个批次并写入 Redis
// 1. 解析事件
// 2. 生成批次
// 3. 写入 Redis
// 4. 提交 Offset
func (kbc *KafkaBatchConsumer) Handle(ctx context.Context, msg xkafka.Message) error {
	kbc.logger.Infof("[kafka批消费者] 开始处理消息 topic=%s partition=%d offset=%d value_len=%d",
		msg.Topic, msg.Partition, msg.Offset, len(msg.Value))

	// 第一步：解析消息
	event, err := kbc.decodeMessage(msg)
	if err != nil {
		kbc.logger.Errorf("[kafka批消费者] 消息解析失败 topic=%s partition=%d offset=%d: %v",
			msg.Topic, msg.Partition, msg.Offset, err)
		// 解析失败，消息无法恢复，跳过（提交 Offset）
		return nil
	}

	// 设置事件的 Kafka 元数据和周期键
	event.KafkaID = fmt.Sprintf("%s_%d_%d", msg.Topic, msg.Partition, msg.Offset)
	event.Topic = msg.Topic
	event.Partition = msg.Partition
	event.Offset = msg.Offset

	// 计算周期键
	taskCode := constants.EventTypeToTaskCode[event.EventType]
	if taskCode == "" {
		kbc.logger.Errorf("[kafka批消费者] 未知的事件类型，跳过处理: event_type=%s, operator_code=%s, kafka_id=%s, trace_id=%s",
			event.EventType, event.OperatorCode, event.KafkaID, event.TraceID)
		return nil
	}
	event.TaskCode = taskCode

	taskConfig, err := kbc.redisQueue.GetTaskConfig(ctx, event.OperatorCode, taskCode)
	if err != nil {
		kbc.logger.Errorf("[kafka批消费者] 获取任务配置失败: %v", err)
		// 任务配置查询失败，重试
		return err
	}

	event.PeriodKey = utils.CalculatePeriodKey(taskConfig, event.OccurredAt)
	event.MaxRetry = constants.MaxTaskRetries

	// 立即写入 Redis
	if err := kbc.redisQueue.DispatchTaskEventToBatch(ctx, event); err != nil {
		kbc.logger.Errorf("[kafka批消费者] 写入批次到redis失败: %v", err)
		return fmt.Errorf("写入批次失败: %w", err)
	}

	kbc.logger.Debugf("[kafka批消费者] 消息已处理并写入redis topic=%s event_type=%s member_id=%d trace_id=%s",
		msg.Topic, event.EventType, event.MemberID, event.TraceID)

	return nil
}

// Flush 刷新所有待处理的批次
// 简化设计：由于每条消息都立即写入 Redis，无需缓冲刷新
func (kbc *KafkaBatchConsumer) Flush(ctx context.Context) error {
	kbc.logger.Debugf("[kafka批消费者] flush已调用，没有待处理的批次")
	return nil
}

// decodeMessage 解析 Kafka 消息为事件
func (kbc *KafkaBatchConsumer) decodeMessage(msg xkafka.Message) (*types.Event, error) {
	// 第一步：解析 Kafka 消息为 types.Payload
	payload, err := xkafka.DecodePayload([]byte(msg.Value))
	if err != nil {
		return nil, fmt.Errorf("payload解析失败: %w", err)
	}

	// 第二步：将 payload.Data 解析为 types.Event
	var event types.Event
	if err := json.Unmarshal([]byte(payload.Data), &event); err != nil {
		return nil, fmt.Errorf("事件解析失败: %w", err)
	}

	return &event, nil
}
