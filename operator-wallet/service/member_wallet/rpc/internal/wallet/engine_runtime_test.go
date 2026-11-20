package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// engineRuntimeTestProducer 模拟钱包运行时使用的 Kafka Producer
type engineRuntimeTestProducer struct {
	publishErr error // 模拟 Kafka 写入结果

	mu         sync.Mutex // 保护关闭次数
	closeCount int        // Producer 关闭次数
	closeOnce  sync.Once  // 确保关闭信号只发送一次
	closed     chan struct{}
}

// publish 模拟 Kafka 事件写入
func (p *engineRuntimeTestProducer) publish(_ context.Context, _ int32, _ walletEvent) error {
	return p.publishErr
}

// close 记录 Producer 关闭
func (p *engineRuntimeTestProducer) close() {
	p.mu.Lock()
	p.closeCount++
	p.mu.Unlock()

	p.closeOnce.Do(func() {
		close(p.closed)
	})
}

// closeCalls 返回 Producer 关闭次数
func (p *engineRuntimeTestProducer) closeCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.closeCount
}

// newEngineRuntimeTest 创建使用模拟 Producer 的钱包运行时
func newEngineRuntimeTest(t *testing.T, handler actorHandler, producer *engineRuntimeTestProducer, expectedCloseErr ...error) *Engine {
	t.Helper()

	core, err := newEngine(2, 1, 8, handler)
	if err != nil {
		t.Fatal(err)
	}

	e := &Engine{
		core:     core,
		producer: producer,
	}

	// 测试结束时确保 Engine 和 Worker 全部退出
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.Close(ctx); err != nil && !isOnlyExpectedError(err, expectedCloseErr) {
			t.Errorf("close wallet engine: %v", err)
		}
	})

	return e
}

// engineRuntimeTestState 创建有效的测试钱包状态
func engineRuntimeTestState() State {
	return State{
		WalletID:       1,
		WalletGlobalID: uuid.New(),
		MemberGlobalID: uuid.New(),
		OperatorCode:   "test-operator",
		CurrencyCode:   "CNY",
		Balance:        500,
		Version:        1,
	}
}

// engineRuntimeTestParams 创建有效的测试资金操作参数
func engineRuntimeTestParams() RecordParams {
	return RecordParams{
		OperationID:    uuid.New(),
		BusinessType:   enum.BusinessTypeDeposit,
		BusinessAction: enum.BusinessActionDepositCredit,
		BusinessNo:     "D10001",
		Changes: []BalanceChange{
			{
				BalanceType:  enum.BalanceTypeAvailable,
				ChangeAmount: 100,
			},
		},
	}
}

// waitEngineRuntimeClosed 等待 Engine 进入关闭状态
func waitEngineRuntimeClosed(t *testing.T, e *Engine) {
	t.Helper()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	for {
		e.core.mu.RLock()
		closed := e.core.closed
		e.core.mu.RUnlock()

		if closed {
			return
		}

		select {
		case <-ticker.C:
		case <-timer.C:
			t.Fatal("wallet engine did not enter closed state")
		}
	}
}

// TestEngineRuntimeCloseWaitsForPendingOperation 验证关闭 Engine 时, 必须等待正在处理的资金操作结束后才能关闭 Producer
func TestEngineRuntimeCloseWaitsForPendingOperation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	var releaseOnce sync.Once
	releaseHandler := func() {
		releaseOnce.Do(func() {
			close(release)
		})
	}

	producer := &engineRuntimeTestProducer{
		closed: make(chan struct{}),
	}

	handler := func(state *State, message any) {
		m := message.(operationMessage)

		close(started)
		<-release

		m.reply <- operationReply{
			result: OperationResult{
				OperationID:    m.params.OperationID,
				WalletGlobalID: state.WalletGlobalID,
			},
		}
	}

	e := newEngineRuntimeTest(t, handler, producer)

	// 确保测试退出时能够释放阻塞中的 Handler
	t.Cleanup(releaseHandler)

	state := engineRuntimeTestState()
	if err := e.RegisterWallet(state); err != nil {
		t.Fatal(err)
	}

	submitDone := make(chan error, 1)

	go func() {
		_, err := e.Submit(
			context.Background(),
			state.WalletGlobalID,
			engineRuntimeTestParams(),
		)
		submitDone <- err
	}()

	// 等待 Actor 开始处理资金操作
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("wallet operation handler did not start")
	}

	closeDone := make(chan error, 1)

	go func() {
		closeDone <- e.Close(context.Background())
	}()

	// 确认 Engine 已经开始关闭
	waitEngineRuntimeClosed(t, e)

	// Handler 尚未结束时, Producer 不允许关闭
	select {
	case <-producer.closed:
		t.Fatal("producer closed before pending operation completed")
	default:
	}

	// Close 不能提前返回
	select {
	case err := <-closeDone:
		t.Fatalf("engine close returned before handler completed: %v", err)
	default:
	}

	// 允许资金操作结束
	releaseHandler()

	select {
	case err := <-submitDone:
		if err != nil {
			t.Fatalf("submit operation: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wallet operation did not finish")
	}

	// 等待 Engine 完成关闭
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("close wallet engine: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wallet engine close did not finish")
	}

	if producer.closeCalls() != 1 {
		t.Fatalf("expected producer close once, got %d", producer.closeCalls())
	}
}

// TestEngineRuntimeCloseCanRetryAfterTimeout 验证 Close 等待超时后, 使用新的 Context 可以继续完成关闭
func TestEngineRuntimeCloseCanRetryAfterTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	var releaseOnce sync.Once
	releaseHandler := func() {
		releaseOnce.Do(func() {
			close(release)
		})
	}

	producer := &engineRuntimeTestProducer{
		closed: make(chan struct{}),
	}

	handler := func(state *State, message any) {
		m := message.(operationMessage)

		close(started)
		<-release

		m.reply <- operationReply{
			result: OperationResult{
				OperationID:    m.params.OperationID,
				WalletGlobalID: state.WalletGlobalID,
			},
		}
	}

	e := newEngineRuntimeTest(t, handler, producer)

	t.Cleanup(releaseHandler)

	state := engineRuntimeTestState()
	if err := e.RegisterWallet(state); err != nil {
		t.Fatal(err)
	}

	submitDone := make(chan error, 1)

	go func() {
		_, err := e.Submit(
			context.Background(),
			state.WalletGlobalID,
			engineRuntimeTestParams(),
		)
		submitDone <- err
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("wallet operation handler did not start")
	}

	// 第一次关闭只等待有限时间
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	err := e.Close(ctx)
	cancel()

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	// 等待超时不能提前关闭 Producer
	if producer.closeCalls() != 0 {
		t.Fatalf("producer closed before pending operation completed: %d", producer.closeCalls())
	}

	select {
	case <-producer.closed:
		t.Fatal("producer closed after close timeout")
	default:
	}

	// 释放正在执行的资金操作
	releaseHandler()

	select {
	case err := <-submitDone:
		if err != nil {
			t.Fatalf("submit operation: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wallet operation did not finish")
	}

	// 使用新的 Context 继续关闭
	retryCtx, retryCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer retryCancel()

	if err := e.Close(retryCtx); err != nil {
		t.Fatalf("retry close wallet engine: %v", err)
	}

	if producer.closeCalls() != 1 {
		t.Fatalf("expected producer close once, got %d", producer.closeCalls())
	}
}

// TestEngineRuntimeCloseAfterFailedActor 验证存在故障 Actor 时, Engine 仍然等待全部资源退出并关闭 Producer
func TestEngineRuntimeCloseAfterFailedActor(t *testing.T) {
	uncertainErr := fmt.Errorf("%w: test kafka timeout", ErrKafkaWriteUncertain)

	producer := &engineRuntimeTestProducer{
		publishErr: uncertainErr,
		closed:     make(chan struct{}),
	}

	handler, err := newOperationHandler(context.Background(), producer, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	e := newEngineRuntimeTest(
		t,
		handler,
		producer,
		ErrWalletFailed,
		ErrKafkaWriteUncertain,
	)

	state := engineRuntimeTestState()
	if err := e.RegisterWallet(state); err != nil {
		t.Fatal(err)
	}

	// 模拟 Kafka 写入结果不明确, 触发 Actor 不可恢复故障
	_, err = e.Submit(
		context.Background(),
		state.WalletGlobalID,
		engineRuntimeTestParams(),
	)
	if !errors.Is(err, ErrKafkaWriteUncertain) {
		t.Fatalf("expected ErrKafkaWriteUncertain, got %v", err)
	}

	// Actor 故障不能阻止 Engine 关闭其他资源
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	closeErr := e.Close(ctx)

	if !errors.Is(closeErr, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", closeErr)
	}

	if !errors.Is(closeErr, ErrKafkaWriteUncertain) {
		t.Fatalf("expected ErrKafkaWriteUncertain, got %v", closeErr)
	}

	// Producer 必须在 Worker Pool 完全退出后关闭
	select {
	case <-producer.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("producer was not closed after failed actor exited")
	}

	if producer.closeCalls() != 1 {
		t.Fatalf("expected producer close once, got %d", producer.closeCalls())
	}
}

// TestEngineRuntimeCloseConcurrent 验证多个 Goroutine 并发关闭 Engine 时, Producer 只关闭一次
func TestEngineRuntimeCloseConcurrent(t *testing.T) {
	producer := &engineRuntimeTestProducer{
		closed: make(chan struct{}),
	}

	e := newEngineRuntimeTest(
		t,
		func(_ *State, _ any) {},
		producer,
	)

	state := engineRuntimeTestState()
	if err := e.RegisterWallet(state); err != nil {
		t.Fatal(err)
	}

	const closeCount = 16

	start := make(chan struct{})
	results := make(chan error, closeCount)

	// 同时触发多个关闭请求
	for range closeCount {
		go func() {
			<-start

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			results <- e.Close(ctx)
		}()
	}

	close(start)

	for range closeCount {
		select {
		case err := <-results:
			if err != nil {
				t.Fatalf("concurrent close wallet engine: %v", err)
			}

		case <-time.After(10 * time.Second):
			t.Fatal("concurrent engine close did not finish")
		}
	}

	// sync.Once 必须保证 Producer 只关闭一次
	if producer.closeCalls() != 1 {
		t.Fatalf("expected producer close once, got %d", producer.closeCalls())
	}

	select {
	case <-producer.closed:
	default:
		t.Fatal("producer was not closed")
	}
}
