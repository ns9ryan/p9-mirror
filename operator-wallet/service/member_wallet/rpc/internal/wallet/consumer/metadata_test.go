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

// TestValidateEmptyTopic 验证首次初始化只允许全新空 Topic
func TestValidateEmptyTopic(t *testing.T) {
	const topic = "test-wallet-empty-topic"

	cluster, err := kfake.NewCluster(
		kfake.SeedTopics(2, topic),
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

	ctx := context.Background()

	// 全新 Topic 的所有 Partition 均为空, 允许初始化
	count, err := ValidateEmptyTopic(ctx, client, topic)
	if err != nil {
		t.Fatalf("validate empty topic: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 partitions, got %d", count)
	}

	// 任意 Partition 写入过消息后, 必须拒绝首次初始化
	err = client.ProduceSync(ctx, &kgo.Record{
		Topic: topic,
		Value: []byte("test-event"),
	}).FirstErr()
	if err != nil {
		t.Fatalf("produce test record: %v", err)
	}

	if _, err := ValidateEmptyTopic(ctx, client, topic); err == nil {
		t.Fatal("expected non-empty topic error")
	}
}
