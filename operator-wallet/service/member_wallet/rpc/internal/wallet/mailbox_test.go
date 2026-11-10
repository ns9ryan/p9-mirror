package wallet

import (
	"context"
	"errors"
	"testing"
)

// TestMailboxFIFO 验证消息队列按 FIFO 顺序消费
func TestMailboxFIFO(t *testing.T) {
	m, err := newMailbox(3)
	if err != nil {
		t.Fatal(err)
	}
	defer m.close()

	ctx := context.Background()
	messages := []string{"first", "second", "third"}

	// 按顺序放入消息
	for _, message := range messages {
		if err := m.enqueue(ctx, message); err != nil {
			t.Fatal(err)
		}
	}

	// 按顺序取出消息
	for _, want := range messages {
		got, err := m.dequeue(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

// TestMailboxFull 验证队列已满时返回错误
func TestMailboxFull(t *testing.T) {
	m, err := newMailbox(1)
	if err != nil {
		t.Fatal(err)
	}
	defer m.close()

	ctx := context.Background()

	if err := m.enqueue(ctx, "first"); err != nil {
		t.Fatal(err)
	}

	if err := m.enqueue(ctx, "second"); !errors.Is(err, ErrMailboxFull) {
		t.Fatalf("expected ErrMailboxFull, got %v", err)
	}
}

// TestMailboxClose 验证队列关闭后的消息处理行为
func TestMailboxClose(t *testing.T) {
	m, err := newMailbox(2)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	if err := m.enqueue(ctx, "pending"); err != nil {
		t.Fatal(err)
	}

	m.close()
	m.close() // 重复关闭不应 panic

	if err := m.enqueue(ctx, "new"); !errors.Is(err, ErrMailboxClosed) {
		t.Fatalf("expected ErrMailboxClosed, got %v", err)
	}

	// 关闭后仍允许消费已经入队的消息
	got, err := m.dequeue(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if got != "pending" {
		t.Fatalf("expected pending, got %v", got)
	}

	if _, err := m.dequeue(ctx); !errors.Is(err, ErrMailboxClosed) {
		t.Fatalf("expected ErrMailboxClosed, got %v", err)
	}
}

// TestMailboxDequeueContextCanceled 验证消息消费支持 Context 取消
func TestMailboxDequeueContextCanceled(t *testing.T) {
	m, err := newMailbox(1)
	if err != nil {
		t.Fatal(err)
	}
	defer m.close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := m.dequeue(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestNewMailboxInvalidCapacity 验证创建队列时拒绝无效容量
func TestNewMailboxInvalidCapacity(t *testing.T) {
	for _, capacity := range []int{0, -1} {
		if _, err := newMailbox(capacity); err == nil {
			t.Errorf("expected error for capacity %d", capacity)
		}
	}
}
