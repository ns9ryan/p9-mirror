package consumer

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// readTopicLogOffsets 读取 Topic 的 Partition 数量及各分区的日志 Offset 范围
func readTopicLogOffsets(ctx context.Context, client *kgo.Client, topic string) (int, map[int32]int64, map[int32]int64, error) {
	// kadm 提供 Kafka 管理接口, 复用现有 kgo.Client 连接
	admin := kadm.NewClient(client)

	// 获取 Topic 元数据, 包括各 Partition 的编号和状态
	topics, err := admin.ListTopics(ctx, topic)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("list Kafka topic %s: %w", topic, err)
	}

	detail, ok := topics[topic]
	if !ok {
		return 0, nil, nil, fmt.Errorf("Kafka topic not found: %s", topic)
	}
	if detail.Err != nil {
		return 0, nil, nil, fmt.Errorf("describe Kafka topic %s: %w", topic, detail.Err)
	}

	partitionCount := len(detail.Partitions)
	if partitionCount == 0 {
		return 0, nil, nil, fmt.Errorf("Kafka topic has no partitions: %s", topic)
	}

	// Log Start Offset: 分区当前最早保留的消息位置, 更早的消息可能已过期删除
	starts, err := admin.ListStartOffsets(ctx, topic)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("list Kafka start offsets: %w", err)
	}
	if err := starts.Error(); err != nil {
		return 0, nil, nil, fmt.Errorf("Kafka start offset error: %w", err)
	}

	// Log End Offset: 分区下一条消息将使用的位置, 不是最后一条消息的 Offset
	ends, err := admin.ListEndOffsets(ctx, topic)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("list Kafka end offsets: %w", err)
	}
	if err := ends.Error(); err != nil {
		return 0, nil, nil, fmt.Errorf("Kafka end offset error: %w", err)
	}

	logStarts := make(map[int32]int64, partitionCount)
	logEnds := make(map[int32]int64, partitionCount)

	// 按 Topic 实际存在的 Partition 逐一提取 Offset, 避免遗漏分区
	for partition := range detail.Partitions {
		// Lookup 返回指定 Topic 和 Partition 的 Offset 查询结果
		start, ok := starts.Lookup(topic, partition)
		if !ok || start.Err != nil || start.Offset < 0 {
			return 0, nil, nil, fmt.Errorf("invalid Kafka start offset: topic=%s partition=%d", topic, partition)
		}

		end, ok := ends.Lookup(topic, partition)
		if !ok || end.Err != nil || end.Offset < start.Offset {
			return 0, nil, nil, fmt.Errorf("invalid Kafka end offset: topic=%s partition=%d", topic, partition)
		}

		logStarts[partition] = start.Offset
		logEnds[partition] = end.Offset
	}

	return partitionCount, logStarts, logEnds, nil
}

// validateEmptyTopic 确认全部 Partition 尚未写入过消息, 返回实际分区数量
func validateEmptyTopic(ctx context.Context, client *kgo.Client, topic string) (int, error) {
	partitionCount, starts, ends, err := readTopicLogOffsets(ctx, client, topic)
	if err != nil {
		return 0, err
	}

	for partition := range int32(partitionCount) {
		// 首次初始化只允许从全新空日志的 Offset 0 开始
		if starts[partition] != 0 || ends[partition] != 0 {
			return 0, fmt.Errorf(
				"wallet checkpoint: topic is not empty: topic=%s partition=%d start=%d end=%d",
				topic, partition, starts[partition], ends[partition],
			)
		}
	}

	return partitionCount, nil
}
