package consumer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
)

// TestBuildPartitionOffsets 验证固定 Partition 的消费起始位置
func TestBuildPartitionOffsets(t *testing.T) {
	const topic = "p9-wallet-events"

	tests := []struct {
		name        string
		count       int
		checkpoints map[int32]int64
		logStarts   map[int32]int64
		logEnds     map[int32]int64
		wantErr     bool
	}{
		{
			name:        "valid offsets",
			count:       2,
			checkpoints: map[int32]int64{0: 100, 1: 200},
			logStarts:   map[int32]int64{0: 50, 1: 100},
			logEnds:     map[int32]int64{0: 150, 1: 250},
		},
		{
			name:        "checkpoint at log end",
			count:       1,
			checkpoints: map[int32]int64{0: 100},
			logStarts:   map[int32]int64{0: 50},
			logEnds:     map[int32]int64{0: 100},
		},
		{
			name:        "empty partition",
			count:       1,
			checkpoints: map[int32]int64{0: 0},
			logStarts:   map[int32]int64{0: 0},
			logEnds:     map[int32]int64{0: 0},
		},
		{
			name:        "missing checkpoint",
			count:       2,
			checkpoints: map[int32]int64{0: 100},
			logStarts:   map[int32]int64{0: 0, 1: 0},
			logEnds:     map[int32]int64{0: 200, 1: 200},
			wantErr:     true,
		},
		{
			name:        "checkpoint before log start",
			count:       1,
			checkpoints: map[int32]int64{0: 49},
			logStarts:   map[int32]int64{0: 50},
			logEnds:     map[int32]int64{0: 100},
			wantErr:     true,
		},
		{
			name:        "checkpoint after log end",
			count:       1,
			checkpoints: map[int32]int64{0: 101},
			logStarts:   map[int32]int64{0: 50},
			logEnds:     map[int32]int64{0: 100},
			wantErr:     true,
		},
		{
			name:        "invalid log range",
			count:       1,
			checkpoints: map[int32]int64{0: 50},
			logStarts:   map[int32]int64{0: 100},
			logEnds:     map[int32]int64{0: 50},
			wantErr:     true,
		},
		{
			name:        "negative log start",
			count:       1,
			checkpoints: map[int32]int64{0: 0},
			logStarts:   map[int32]int64{0: -1},
			logEnds:     map[int32]int64{0: 100},
			wantErr:     true,
		},
		{
			name:        "unexpected partition",
			count:       1,
			checkpoints: map[int32]int64{1: 0},
			logStarts:   map[int32]int64{1: 0},
			logEnds:     map[int32]int64{1: 100},
			wantErr:     true,
		},
		{
			name:    "invalid partition count",
			count:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPartitionOffsets(
				topic,
				tt.count,
				tt.checkpoints,
				tt.logStarts,
				tt.logEnds,
			)

			if tt.wantErr {
				if !errors.Is(err, errInvalidConsumeOffsets) {
					t.Fatalf("expected errInvalidConsumeOffsets, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("build partition offsets: %v", err)
			}

			want := make(map[int32]kgo.Offset, tt.count)
			for partition, offset := range tt.checkpoints {
				want[partition] = kgo.NewOffset().At(offset)
			}

			if !reflect.DeepEqual(got, map[string]map[int32]kgo.Offset{topic: want}) {
				t.Fatalf("unexpected partition offsets: got=%v", got)
			}
		})
	}
}

// TestBuildPartitionOffsetsEmptyTopic 验证空 Topic 不允许启动消费
func TestBuildPartitionOffsetsEmptyTopic(t *testing.T) {
	_, err := buildPartitionOffsets(
		"",
		1,
		map[int32]int64{0: 0},
		map[int32]int64{0: 0},
		map[int32]int64{0: 0},
	)
	if !errors.Is(err, errInvalidConsumeOffsets) {
		t.Fatalf("expected errInvalidConsumeOffsets, got %v", err)
	}
}
