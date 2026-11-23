package consumer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// newTestPartitionWorker 创建测试 Worker, 并注册退出清理
func newTestPartitionWorker(t *testing.T, partition int32, capacity int, handler Handler) *partitionWorker {
	t.Helper()

	w, err := newPartitionWorker(context.Background(), partition, capacity, handler)
	if err != nil {
		t.Fatalf("create partition worker: %v", err)
	}

	t.Cleanup(func() {
		w.stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := w.wait(ctx); errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("partition worker did not stop: %v", err)
		}
	})

	return w
}

// testPartitionRecord 构造指定 Partition 和 Offset 的消息
func testPartitionRecord(partition int32, offset int64) *kgo.Record {
	return &kgo.Record{
		Topic:     testTopic,
		Partition: partition,
		Offset:    offset,
	}
}

// waitForWorker 等待 Worker 退出, 防止测试无限阻塞
func waitForWorker(t *testing.T, w *partitionWorker) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := w.wait(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timed out waiting for partition worker")
	}

	return err
}

// TestPartitionWorkerInvalidParams 验证创建参数和消息校验
func TestPartitionWorkerInvalidParams(t *testing.T) {
	handler := func(context.Context, *kgo.Record) error {
		return nil
	}

	tests := []struct {
		name      string
		ctx       context.Context
		partition int32
		capacity  int
		handler   Handler
		wantErr   error
	}{
		{"nil context", nil, 0, 1, handler, errNilContext},
		{"invalid partition", context.Background(), -1, 1, handler, errInvalidPartition},
		{"invalid capacity", context.Background(), 0, 0, handler, errInvalidWorkerCapacity},
		{"nil handler", context.Background(), 0, 1, nil, errNilHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newPartitionWorker(tt.ctx, tt.partition, tt.capacity, tt.handler)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	w := newTestPartitionWorker(t, 0, 1, handler)

	if err := w.enqueue(nil, testPartitionRecord(0, 0)); !errors.Is(err, errNilContext) {
		t.Fatalf("expected errNilContext, got %v", err)
	}

	if err := w.enqueue(context.Background(), nil); !errors.Is(err, errNilRecord) {
		t.Fatalf("expected errNilRecord, got %v", err)
	}

	if err := w.enqueue(context.Background(), testPartitionRecord(1, 0)); !errors.Is(err, errPartitionMismatch) {
		t.Fatalf("expected errPartitionMismatch, got %v", err)
	}
}

// TestPartitionWorkerFIFO 验证同一 Partition 按入队顺序处理
func TestPartitionWorkerFIFO(t *testing.T) {
	var mu sync.Mutex
	var offsets []int64

	w := newTestPartitionWorker(t, 3, 16, func(_ context.Context, record *kgo.Record) error {
		mu.Lock()
		offsets = append(offsets, record.Offset)
		mu.Unlock()
		return nil
	})

	const count = 10

	for i := range count {
		if err := w.enqueue(context.Background(), testPartitionRecord(3, int64(i))); err != nil {
			t.Fatalf("enqueue offset %d: %v", i, err)
		}
	}

	w.stop()

	if err := waitForWorker(t, w); err != nil {
		t.Fatalf("wait worker: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(offsets) != count {
		t.Fatalf("expected %d records, got %d", count, len(offsets))
	}

	for i, offset := range offsets {
		if offset != int64(i) {
			t.Fatalf("record %d: expected offset %d, got %d", i, i, offset)
		}
	}
}

// TestPartitionWorkerParallel 验证不同 Partition 可以并行处理
func TestPartitionWorkerParallel(t *testing.T) {
	started := make(chan int32, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once

	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
	})

	handler := func(_ context.Context, record *kgo.Record) error {
		started <- record.Partition
		<-release
		return nil
	}

	w0 := newTestPartitionWorker(t, 0, 1, handler)
	w1 := newTestPartitionWorker(t, 1, 1, handler)

	if err := w0.enqueue(context.Background(), testPartitionRecord(0, 0)); err != nil {
		t.Fatal(err)
	}
	if err := w1.enqueue(context.Background(), testPartitionRecord(1, 0)); err != nil {
		t.Fatal(err)
	}

	// 两个 Handler 都应在释放前启动, 证明没有互相阻塞
	received := make(map[int32]bool)

	for range 2 {
		select {
		case partition := <-started:
			received[partition] = true
		case <-time.After(5 * time.Second):
			t.Fatal("partition workers did not run concurrently")
		}
	}

	if !received[0] || !received[1] {
		t.Fatalf("unexpected started partitions: %v", received)
	}

	releaseOnce.Do(func() { close(release) })

	w0.stop()
	w1.stop()

	if err := waitForWorker(t, w0); err != nil {
		t.Fatal(err)
	}
	if err := waitForWorker(t, w1); err != nil {
		t.Fatal(err)
	}
}

// TestPartitionWorkerBackpressure 验证队列满时阻塞入队
func TestPartitionWorkerBackpressure(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once

	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
	})

	w := newTestPartitionWorker(t, 0, 1, func(_ context.Context, record *kgo.Record) error {
		if record.Offset == 0 {
			close(started)
			<-release
		}
		return nil
	})

	// Offset 0 正在执行, Offset 1 占满队列
	if err := w.enqueue(context.Background(), testPartitionRecord(0, 0)); err != nil {
		t.Fatal(err)
	}

	waitForSignal(t, started)

	if err := w.enqueue(context.Background(), testPartitionRecord(0, 1)); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	enqueued := make(chan error, 1)

	go func() {
		enqueued <- w.enqueue(ctx, testPartitionRecord(0, 2))
	}()

	// 队列已满, 第三条消息不能立即入队
	select {
	case err := <-enqueued:
		t.Fatalf("enqueue should be blocked, got %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	// 释放 Handler 后, Worker 继续取消息并腾出队列容量
	releaseOnce.Do(func() { close(release) })

	select {
	case err := <-enqueued:
		if err != nil {
			t.Fatalf("enqueue after capacity released: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("enqueue did not resume")
	}

	w.stop()

	if err := waitForWorker(t, w); err != nil {
		t.Fatal(err)
	}
}

// TestPartitionWorkerStopDrainsPending 验证正常停止后排空已入队消息
func TestPartitionWorkerStopDrainsPending(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	var processed atomic.Int32

	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
	})

	w := newTestPartitionWorker(t, 0, 4, func(_ context.Context, record *kgo.Record) error {
		if record.Offset == 0 {
			close(started)
			<-release
		}

		processed.Add(1)
		return nil
	})

	for i := range 4 {
		if err := w.enqueue(context.Background(), testPartitionRecord(0, int64(i))); err != nil {
			t.Fatal(err)
		}
	}

	waitForSignal(t, started)

	w.stop()
	w.stop()

	if err := w.enqueue(context.Background(), testPartitionRecord(0, 4)); !errors.Is(err, errPartitionWorkerStopped) {
		t.Fatalf("expected errPartitionWorkerStopped, got %v", err)
	}

	releaseOnce.Do(func() { close(release) })

	if err := waitForWorker(t, w); err != nil {
		t.Fatal(err)
	}

	if got := processed.Load(); got != 4 {
		t.Fatalf("expected 4 processed records, got %d", got)
	}
}

// TestPartitionWorkerHandlerError 验证 Handler 失败后停止处理后续消息
func TestPartitionWorkerHandlerError(t *testing.T) {
	expectedErr := errors.New("invalid wallet event")
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	var processed atomic.Int32

	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
	})

	w := newTestPartitionWorker(t, 5, 2, func(_ context.Context, record *kgo.Record) error {
		if record.Offset == 100 {
			close(started)
			<-release
			return expectedErr
		}

		processed.Add(1)
		return nil
	})

	if err := w.enqueue(context.Background(), testPartitionRecord(5, 100)); err != nil {
		t.Fatal(err)
	}

	waitForSignal(t, started)

	// 后续消息已入队, 但失败后不能继续处理
	if err := w.enqueue(context.Background(), testPartitionRecord(5, 101)); err != nil {
		t.Fatal(err)
	}

	releaseOnce.Do(func() { close(release) })

	if err := waitForWorker(t, w); !errors.Is(err, expectedErr) {
		t.Fatalf("expected handler error, got %v", err)
	}

	if got := processed.Load(); got != 0 {
		t.Fatalf("records after failed offset must not be processed, got %d", got)
	}

	if err := w.enqueue(context.Background(), testPartitionRecord(5, 102)); !errors.Is(err, errPartitionWorkerStopped) {
		t.Fatalf("expected errPartitionWorkerStopped, got %v", err)
	}
}

// TestPartitionWorkerContextCanceled 验证执行 Context 取消后退出
func TestPartitionWorkerContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})

	w, err := newPartitionWorker(ctx, 0, 1, func(ctx context.Context, _ *kgo.Record) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cancel()
		waitForWorker(t, w)
	})

	if err := w.enqueue(context.Background(), testPartitionRecord(0, 0)); err != nil {
		t.Fatal(err)
	}

	waitForSignal(t, started)
	cancel()

	if err := waitForWorker(t, w); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestPartitionWorkerWaitTimeout 验证 Wait 超时不影响 Worker 继续执行
func TestPartitionWorkerWaitTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once

	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
	})

	w := newTestPartitionWorker(t, 0, 1, func(_ context.Context, _ *kgo.Record) error {
		close(started)
		<-release
		return nil
	})

	if err := w.enqueue(context.Background(), testPartitionRecord(0, 0)); err != nil {
		t.Fatal(err)
	}

	waitForSignal(t, started)
	w.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := w.wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	releaseOnce.Do(func() { close(release) })

	if err := waitForWorker(t, w); err != nil {
		t.Fatal(err)
	}

	// Worker 已退出, 重复 Wait 应立即返回原结果
	if err := waitForWorker(t, w); err != nil {
		t.Fatal(err)
	}
}

// TestPartitionWorkerConcurrentStop 验证并发停止和入队不会发生竞争
func TestPartitionWorkerConcurrentStop(t *testing.T) {
	var accepted atomic.Int32
	var processed atomic.Int32

	w := newTestPartitionWorker(t, 0, 64, func(_ context.Context, _ *kgo.Record) error {
		processed.Add(1)
		return nil
	})

	const callers = 32

	var wg sync.WaitGroup
	wg.Add(callers + 1)

	for i := range callers {
		go func(offset int64) {
			defer wg.Done()

			err := w.enqueue(context.Background(), testPartitionRecord(0, offset))
			switch {
			case err == nil:
				accepted.Add(1)
			case errors.Is(err, errPartitionWorkerStopped):
			default:
				t.Errorf("unexpected enqueue error: %v", err)
			}
		}(int64(i))
	}

	go func() {
		defer wg.Done()
		w.stop()
	}()

	wg.Wait()

	if err := waitForWorker(t, w); err != nil {
		t.Fatal(err)
	}

	// 正常停止时, 所有成功入队的消息都必须处理
	if got, want := processed.Load(), accepted.Load(); got != want {
		t.Fatalf("processed=%d, accepted=%d", got, want)
	}
}
