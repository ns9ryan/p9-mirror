package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/twmb/franz-go/pkg/kfake"
)

// TestNewFromCheckpoints 验证根据 Kafka 日志范围和数据库进度创建 Consumer
func TestNewFromCheckpoints(t *testing.T) {
	const topic = "test-wallet-bootstrap"

	cluster, err := kfake.NewCluster(
		kfake.SeedTopics(2, topic),
	)
	if err != nil {
		t.Fatalf("create Kafka cluster: %v", err)
	}
	t.Cleanup(cluster.Close)

	brokers := cluster.ListenAddrs()

	tests := []struct {
		name        string
		checkpoints map[int32]int64
		wantErr     bool
	}{
		{
			name:        "valid checkpoints",
			checkpoints: map[int32]int64{0: 0, 1: 0},
		},
		{
			name:        "missing checkpoint",
			checkpoints: map[int32]int64{0: 0},
			wantErr:     true,
		},
		{
			name:        "checkpoint after log end",
			checkpoints: map[int32]int64{0: 1, 1: 0},
			wantErr:     true,
		},
		{
			name:        "unexpected partition",
			checkpoints: map[int32]int64{0: 0, 2: 0},
			wantErr:     true,
		},
		{
			name:    "empty checkpoints",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewFromCheckpoints(
				context.Background(),
				brokers,
				topic,
				tt.checkpoints,
			)

			if tt.wantErr {
				if !errors.Is(err, errInvalidConsumeOffsets) {
					t.Fatalf("expected invalid consume offsets, got %v", err)
				}
				if c != nil {
					c.Close()
					t.Fatal("expected nil consumer")
				}
				return
			}

			if err != nil {
				t.Fatalf("create consumer: %v", err)
			}
			if c == nil {
				t.Fatal("expected non-nil consumer")
			}

			c.Close()
		})
	}
}

// TestNewFromCheckpointsCanceled 验证取消的 Context 不允许创建 Consumer
func TestNewFromCheckpointsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c, err := NewFromCheckpoints(
		ctx,
		[]string{"localhost:9092"},
		"test-wallet-bootstrap",
		map[int32]int64{0: 0},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if c != nil {
		c.Close()
		t.Fatal("expected nil consumer")
	}
}

// TestNewFromCheckpointsNilContext 验证拒绝空 Context
func TestNewFromCheckpointsNilContext(t *testing.T) {
	c, err := NewFromCheckpoints(
		nil,
		[]string{"localhost:9092"},
		"test-wallet-bootstrap",
		map[int32]int64{0: 0},
	)

	if !errors.Is(err, errNilContext) {
		t.Fatalf("expected errNilContext, got %v", err)
	}
	if c != nil {
		c.Close()
		t.Fatal("expected nil consumer")
	}
}
