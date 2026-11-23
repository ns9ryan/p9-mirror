package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

const testTopic = "test-wallet-events"

// newTestConsumer 创建使用内存 Kafka 集群的测试 Consumer
func newTestConsumer(t *testing.T) (*Consumer, *kgo.Client, *kfake.Cluster) {
	t.Helper()

	cluster, err := kfake.NewCluster(
		kfake.SeedTopics(1, testTopic),
	)
	if err != nil {
		t.Fatalf("create fake Kafka cluster: %v", err)
	}

	t.Cleanup(cluster.Close)

	c, err := New(Config{
		Brokers: cluster.ListenAddrs(),
		Topic:   testTopic,
		Partitions: map[int32]kgo.Offset{
			0: kgo.NewOffset().At(0),
		},
	})
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_ = c.Stop(ctx)
		c.Close()
	})

	producer, err := kgo.NewClient(
		kgo.SeedBrokers(cluster.ListenAddrs()...),
	)
	if err != nil {
		t.Fatalf("create test producer: %v", err)
	}

	t.Cleanup(producer.Close)

	return c, producer, cluster
}

// produceTestRecord 向测试 Topic 写入一条消息
func produceTestRecord(t *testing.T, producer *kgo.Client, value string) {
	t.Helper()

	err := producer.ProduceSync(
		context.Background(),
		&kgo.Record{
			Topic: testTopic,
			Value: []byte(value),
		},
	).FirstErr()
	if err != nil {
		t.Fatalf("produce test record: %v", err)
	}
}

// waitForSignal 等待信号或超时
func waitForSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-ch:
	case <-timer.C:
		t.Fatal("timed out waiting for signal")
	}
}

// TestConsumerStartStop 验证 Consumer 正常启动和停止
func TestConsumerStartStop(t *testing.T) {
	c, _, _ := newTestConsumer(t)

	if err := c.Start(context.Background(), func(context.Context, *kgo.Record) error {
		return nil
	}); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Stop(ctx); err != nil {
		t.Fatalf("stop consumer: %v", err)
	}
}

// TestConsumerStartTwice 验证 Consumer 不允许重复启动
func TestConsumerStartTwice(t *testing.T) {
	c, _, _ := newTestConsumer(t)

	handler := func(context.Context, *kgo.Record) error {
		return nil
	}

	if err := c.Start(context.Background(), handler); err != nil {
		t.Fatalf("first start: %v", err)
	}

	if err := c.Start(context.Background(), handler); !errors.Is(err, errConsumerAlreadyStarted) {
		t.Fatalf("expected errConsumerAlreadyStarted, got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Stop(ctx); err != nil {
		t.Fatalf("stop consumer: %v", err)
	}
}

// TestConsumerContextCanceled 验证父 Context 取消后消费循环退出
func TestConsumerContextCanceled(t *testing.T) {
	c, _, _ := newTestConsumer(t)

	ctx, cancel := context.WithCancel(context.Background())

	if err := c.Start(ctx, func(context.Context, *kgo.Record) error {
		return nil
	}); err != nil {
		cancel()
		t.Fatalf("start consumer: %v", err)
	}

	cancel()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := c.Stop(stopCtx); err != nil {
		t.Fatalf("stop consumer: %v", err)
	}
}

// TestConsumerHandlerError 验证 Handler 错误能够被保存并返回
func TestConsumerHandlerError(t *testing.T) {
	c, producer, _ := newTestConsumer(t)

	produceTestRecord(t, producer, "test-event")

	expectedErr := errors.New("test handler error")
	handlerCalled := make(chan struct{})

	if err := c.Start(context.Background(), func(_ context.Context, record *kgo.Record) error {
		close(handlerCalled)
		return expectedErr
	}); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	waitForSignal(t, handlerCalled)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Stop(ctx); !errors.Is(err, expectedErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

// TestConsumerStopTimeout 验证 Stop 超时后可以再次等待退出
func TestConsumerStopTimeout(t *testing.T) {
	c, producer, _ := newTestConsumer(t)

	produceTestRecord(t, producer, "test-event")

	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})

	if err := c.Start(context.Background(), func(_ context.Context, _ *kgo.Record) error {
		close(handlerStarted)
		<-releaseHandler
		return nil
	}); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	waitForSignal(t, handlerStarted)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := c.Stop(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}

	// 允许 Handler 结束，消费循环才能退出
	close(releaseHandler)

	retryCtx, retryCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer retryCancel()

	if err := c.Stop(retryCtx); err != nil {
		t.Fatalf("retry stop consumer: %v", err)
	}
}

// TestConsumerConcurrentStop 验证多个 Goroutine 同时停止 Consumer
func TestConsumerConcurrentStop(t *testing.T) {
	c, _, _ := newTestConsumer(t)

	if err := c.Start(context.Background(), func(context.Context, *kgo.Record) error {
		return nil
	}); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	const callers = 8

	errs := make(chan error, callers)

	for range callers {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			errs <- c.Stop(ctx)
		}()
	}

	for range callers {
		if err := <-errs; err != nil {
			t.Errorf("stop consumer: %v", err)
		}
	}
}

// TestConsumerClose 验证重复关闭以及关闭后禁止启动
func TestConsumerClose(t *testing.T) {
	c, _, _ := newTestConsumer(t)

	if err := c.Start(context.Background(), func(context.Context, *kgo.Record) error {
		return nil
	}); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Stop(ctx); err != nil {
		t.Fatalf("stop consumer: %v", err)
	}

	c.Close()
	c.Close()

	if err := c.Start(context.Background(), func(context.Context, *kgo.Record) error {
		return nil
	}); !errors.Is(err, errConsumerClosed) {
		t.Fatalf("expected errConsumerClosed, got %v", err)
	}
}

// TestConsumerStopDuringHandler 验证正常停止不会产生虚假的 Handler 错误
func TestConsumerStopDuringHandler(t *testing.T) {
	c, producer, _ := newTestConsumer(t)

	produceTestRecord(t, producer, "first")
	produceTestRecord(t, producer, "second")

	started := make(chan struct{})
	release := make(chan struct{})

	if err := c.Start(context.Background(), func(_ context.Context, record *kgo.Record) error {
		if string(record.Value) != "first" {
			t.Errorf("unexpected record after stop: %s", record.Value)
			return errors.New("unexpected record")
		}

		close(started)
		<-release
		return nil
	}); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	waitForSignal(t, started)

	// 先取消消费 Context, 再允许当前 Handler 返回
	c.mu.Lock()
	cancel := c.cancel
	c.mu.Unlock()

	cancel()
	close(release)

	ctx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := c.Stop(ctx); err != nil {
		t.Fatalf("expected clean stop, got %v", err)
	}
}
