package consumer

import (
	"context"
	"testing"

	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

// TestReadTopicLogOffsets 验证 Kafka 元数据和各分区日志范围
func TestReadTopicLogOffsets(t *testing.T) {
	const (
		topic          = "test-wallet-metadata"
		partitionCount = 2
	)

	cluster, err := kfake.NewCluster(
		kfake.SeedTopics(partitionCount, topic),
	)
	if err != nil {
		t.Fatalf("create Kafka cluster: %v", err)
	}
	t.Cleanup(cluster.Close)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cluster.ListenAddrs()...),
	)
	if err != nil {
		t.Fatalf("create Kafka client: %v", err)
	}
	t.Cleanup(client.Close)

	count, starts, ends, err := readTopicLogOffsets(context.Background(), client, topic)
	if err != nil {
		t.Fatalf("read Kafka topic offsets: %v", err)
	}

	if count != partitionCount {
		t.Fatalf("expected %d partitions, got %d", partitionCount, count)
	}

	// 新建 Topic 尚无消息, 各分区起点和终点都应为 0
	for partition := range int32(partitionCount) {
		start, hasStart := starts[partition]
		end, hasEnd := ends[partition]

		if !hasStart || !hasEnd || start != 0 || end != 0 {
			t.Fatalf("unexpected offsets: partition=%d start=%d end=%d", partition, start, end)
		}
	}
}
