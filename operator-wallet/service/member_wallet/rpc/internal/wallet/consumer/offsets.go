package consumer

import (
	"errors"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

var errInvalidConsumeOffsets = errors.New("wallet consumer: invalid consume offsets")

// buildPartitionOffsets 校验消费进度并生成固定 Partition 订阅配置
func buildPartitionOffsets(
	topic string,
	partitionCount int,
	checkpoints map[int32]int64,
	logStarts map[int32]int64,
	logEnds map[int32]int64,
) (map[string]map[int32]kgo.Offset, error) {
	if topic == "" || partitionCount <= 0 {
		return nil, errInvalidConsumeOffsets
	}

	// 必须覆盖全部 Partition, 不允许遗漏消费进度
	if len(checkpoints) != partitionCount || len(logStarts) != partitionCount || len(logEnds) != partitionCount {
		return nil, fmt.Errorf(
			"%w: incomplete partitions: expected=%d checkpoints=%d starts=%d ends=%d",
			errInvalidConsumeOffsets,
			partitionCount, len(checkpoints), len(logStarts), len(logEnds),
		)
	}

	partitions := make(map[int32]kgo.Offset, partitionCount)

	for i := range partitionCount {
		partition := int32(i)

		nextOffset, ok := checkpoints[partition]
		if !ok {
			return nil, fmt.Errorf("%w: missing checkpoint: partition=%d", errInvalidConsumeOffsets, partition)
		}

		start, hasStart := logStarts[partition]
		end, hasEnd := logEnds[partition]
		if !hasStart || !hasEnd || start < 0 || end < start {
			return nil, fmt.Errorf(
				"%w: invalid log range: partition=%d start=%d end=%d",
				errInvalidConsumeOffsets, partition, start, end,
			)
		}

		// 小于日志起点说明历史消息可能丢失; 大于日志终点说明进度异常
		if nextOffset < start || nextOffset > end {
			return nil, fmt.Errorf(
				"%w: checkpoint out of range: partition=%d next=%d start=%d end=%d",
				errInvalidConsumeOffsets, partition, nextOffset, start, end,
			)
		}

		partitions[partition] = kgo.NewOffset().At(nextOffset)
	}

	return map[string]map[int32]kgo.Offset{
		topic: partitions,
	}, nil
}
