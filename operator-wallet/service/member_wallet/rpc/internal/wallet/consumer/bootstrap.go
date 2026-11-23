package consumer

import (
	"context"
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

// NewFromCheckpoints 根据数据库消费进度和 Kafka 日志范围创建 Consumer
func NewFromCheckpoints(ctx context.Context, brokers []string, topic string, checkpoints map[int32]int64) (*Consumer, error) {
	if ctx == nil {
		return nil, errNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(brokers) == 0 {
		return nil, errEmptyBrokers
	}
	for _, broker := range brokers {
		if broker == "" || strings.TrimSpace(broker) != broker {
			return nil, fmt.Errorf("wallet consumer: invalid broker address")
		}
	}
	if topic == "" || strings.TrimSpace(topic) != topic {
		return nil, errEmptyTopic
	}

	// 临时客户端仅用于读取 Kafka 元数据, 不启动消息消费
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("create Kafka metadata client: %w", err)
	}
	defer client.Close()

	// 获取 Topic 实际分区数量及各分区可读取的日志范围
	partitionCount, logStarts, logEnds, err := readTopicLogOffsets(ctx, client, topic)
	if err != nil {
		return nil, err
	}

	// 校验 PostgreSQL Checkpoint, 禁止跳过缺失或过期的资金事件
	offsets, err := buildPartitionOffsets(topic, partitionCount, checkpoints, logStarts, logEnds)
	if err != nil {
		return nil, fmt.Errorf("validate wallet consumer checkpoints: %w", err)
	}

	// 从经过校验的 Offset 创建固定 Partition Consumer
	return New(Config{
		Brokers:    brokers,
		Topic:      topic,
		Partitions: offsets[topic],
	})
}
