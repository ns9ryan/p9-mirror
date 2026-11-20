// Package kafka 实现基于 xkafka 的 Kafka 消费者。
// 新版本：使用批次处理，按批次聚合消息并批量写入 Redis，只有成功写入后才提交 Offset
package kafka

import (
	"context"
	"sync"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
	"github.com/zeromicro/go-zero/core/service"
	xkafka "oa.98ent.com/p9/common/xkafka"
	"oa.98ent.com/p9/operator-activity/task/constants"
	taskqueue "oa.98ent.com/p9/operator-activity/task/queue"
)

// 全局批次消费者缓存（按 topic）
var (
	batchConsumers = make(map[string]*KafkaBatchConsumer)
	bcMutex        sync.RWMutex
)

// StartKq 启动 Kafka 消费者
// 订阅所有定义的 topic，消费消息并推送到 RedisQueue
// redisQueue: 接收事件的 Redis 队列
// brokers: Kafka broker 地址列表
func StartKq(ctx context.Context, redisQueue *taskqueue.RedisQueue, brokers []string) []queue.MessageQueue {
	if redisQueue == nil {
		logx.Error("[kq] redis queue is nil")
		return nil
	}
	if len(brokers) == 0 {
		logx.Error("[kq] brokers is empty")
		return nil
	}

	var queues []queue.MessageQueue
	consumerGroup := "operator-activity-task"
	// 为每个 topic 创建独立的消费者
	for _, topic := range constants.AllTopics {
		// 为该 topic 创建批次消费者
		batchConsumer := NewKafkaBatchConsumer(ctx, redisQueue, topic)
		bcMutex.Lock()
		batchConsumers[topic] = batchConsumer
		bcMutex.Unlock()

		conf := kq.KqConf{
			ServiceConf: service.ServiceConf{},
			Brokers:     brokers,
			Group:       consumerGroup,
			Topic:       topic,
			Offset:      "last",
			Consumers:   1,
			Processors:  4,
			ForceCommit: false,
			MinBytes:    1,
			MaxBytes:    10485760, // 10MB
		}

		// 创建消息处理器
		handler := createBatchMessageHandler(batchConsumer)
		q := xkafka.Consume(conf, handler)
		if q != nil {
			queues = append(queues, q)
			logx.Infof("[Kafka消费者] 消费者已启动: topic=%s, group=%s, offset=%s", topic, consumerGroup, conf.Offset)
		} else {
			logx.Errorf("[Kafka消费者] 启动消费者失败: topic=%s", topic)
		}
	}

	logx.Infof("[kq-batch] kafka batch consumer started, total topics subscribed=%d", len(queues))
	return queues
}

// createBatchMessageHandler 创建批次消息处理函数
// 将 Kafka 消息通过 KafkaBatchConsumer 处理（聚合、批量写入 Redis）
func createBatchMessageHandler(batchConsumer *KafkaBatchConsumer) xkafka.Handler {
	return func(ctx context.Context, msg xkafka.Message) error {
		return batchConsumer.Handle(ctx, msg)
	}
}

// StopKq 优雅停止所有消费者
// 在停止之前 Flush 所有待处理的批次
func StopKq(queues []queue.MessageQueue) {
	ctx := context.Background()

	// Flush 所有待处理的批次
	bcMutex.RLock()
	for topic, consumer := range batchConsumers {
		if err := consumer.Flush(ctx); err != nil {
			logx.Errorf("[kq-batch] failed to flush batches for topic=%s: %v", topic, err)
		}
	}
	bcMutex.RUnlock()

	// 停止所有消费者
	xkafka.StopMany(queues)
	logx.Info("[kq-batch] kafka batch consumers stopped")
}
