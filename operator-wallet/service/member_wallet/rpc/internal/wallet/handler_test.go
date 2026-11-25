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

// handlerTestProducer 模拟 Kafka 生产者
type handlerTestProducer struct {
	publishFunc func(context.Context, int32, walletEvent) error
}

func (p *handlerTestProducer) publish(ctx context.Context, partition int32, event walletEvent) error {
	return p.publishFunc(ctx, partition, event)
}

// newHandlerTestEngine 创建使用生产 Handler 的测试 Engine
func newHandlerTestEngine(t *testing.T, producer eventProducer) (*engine, *actor, State) {
	t.Helper()

	handler, err := newOperationHandler(context.Background(), producer, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	e, err := newEngine(2, 1, 8, handler)
	if err != nil {
		t.Fatal(err)
	}

	state := testWalletState()
	state.Balance = 500
	state.Version = 7

	a, err := e.getOrCreate(state)
	if err != nil {
		t.Fatal(err)
	}

	// 故障 Actor 的退出错误是预期结果, 不应影响测试清理
	t.Cleanup(func() {
		e.close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := e.wait(ctx); err != nil && !errors.Is(err, ErrWalletFailed) {
			t.Errorf("wait engine: %v", err)
		}
	})

	return e, a, state
}

// handlerTestParams 创建测试资金操作参数
func handlerTestParams() RecordParams {
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

// waitForPendingMailboxMessage 等待 Actor Mailbox 中出现待处理消息
func waitForPendingMailboxMessage(t *testing.T, a *actor) {
	t.Helper()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for len(a.mailbox.queue) == 0 {
		select {
		case <-ticker.C:
		case <-timeout.C:
			t.Fatal("actor mailbox did not receive pending message")
		}
	}
}

// TestOperationHandlerPublishesAndUpdatesState 验证 Kafka 确认后更新钱包状态并返回结果
func TestOperationHandlerPublishesAndUpdatesState(t *testing.T) {
	published := make(chan walletEvent, 1)

	producer := &handlerTestProducer{
		publishFunc: func(_ context.Context, _ int32, event walletEvent) error {
			published <- event
			return nil
		},
	}

	e, a, initialState := newHandlerTestEngine(t, producer)
	params := handlerTestParams()

	result, err := e.submit(context.Background(), initialState.WalletGlobalID, params)
	if err != nil {
		t.Fatal(err)
	}

	event := <-published

	if event.OperationID != params.OperationID || event.BusinessNo != params.BusinessNo {
		t.Fatalf("unexpected wallet event: %+v", event)
	}

	if result.OperationID != params.OperationID || result.Version != 8 || result.Balance != 600 {
		t.Fatalf("unexpected operation result: %+v", result)
	}

	// Handler 返回结果前必须已经更新正式 State
	if a.state.Balance != 600 || a.state.Version != 8 {
		t.Fatalf("unexpected actor state: %+v", a.state)
	}
}

// TestOperationHandlerFailsActorOnUncertainWrite 验证 Kafka 结果不明确时终止 Actor
func TestOperationHandlerFailsActorOnUncertainWrite(t *testing.T) {
	producer := &handlerTestProducer{
		publishFunc: func(_ context.Context, _ int32, _ walletEvent) error {
			return fmt.Errorf("%w: test timeout", ErrKafkaWriteUncertain)
		},
	}

	e, a, initialState := newHandlerTestEngine(t, producer)
	params := handlerTestParams()

	_, err := e.submit(context.Background(), initialState.WalletGlobalID, params)
	if !errors.Is(err, ErrKafkaWriteUncertain) {
		t.Fatalf("expected ErrKafkaWriteUncertain, got %v", err)
	}

	// 故障发生后, Actor 必须完全退出
	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	// 故障后不能提前修改正式余额和版本
	if a.state != initialState {
		t.Fatalf("wallet state changed before confirmation: %+v", a.state)
	}

	// 故障信息必须保留原始资金事件
	failure, ok := a.failureInfo()
	if !ok {
		t.Fatal("expected wallet failure info")
	}

	if failure.event.OperationID != params.OperationID {
		t.Fatalf("unexpected operation ID: %s", failure.event.OperationID)
	}

	if failure.event.WalletGlobalID != initialState.WalletGlobalID {
		t.Fatal("failure event wallet ID mismatch")
	}

	if failure.event.BusinessNo != params.BusinessNo {
		t.Fatalf("unexpected business number: %s", failure.event.BusinessNo)
	}

	if len(failure.event.Ledgers) != 1 {
		t.Fatalf("expected 1 ledger, got %d", len(failure.event.Ledgers))
	}

	ledger := failure.event.Ledgers[0]
	if ledger.ChangeAmount != 100 ||
		ledger.BalanceBefore != 500 ||
		ledger.BalanceAfter != 600 ||
		ledger.LedgerGlobalID == uuid.Nil {
		t.Fatalf("unexpected failure ledger: %+v", ledger)
	}

	if !errors.Is(failure.cause, ErrKafkaWriteUncertain) {
		t.Fatalf("unexpected failure cause: %v", failure.cause)
	}

	// 验证 Actor 已经进入不可自动恢复的故障状态
	a.mu.Lock()
	failed := a.failed
	a.mu.Unlock()

	if !failed {
		t.Fatal("actor should enter failed state after uncertain Kafka write")
	}

	// 故障后再次提交资金操作必须被拒绝
	_, err = e.submit(context.Background(), initialState.WalletGlobalID, handlerTestParams())
	if !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed on subsequent operation, got %v", err)
	}
}

// TestOperationHandlerDoesNotFailActorOnDefiniteError 验证明确失败不会将 Actor 标记为故障
func TestOperationHandlerDoesNotFailActorOnDefiniteError(t *testing.T) {
	expectedErr := errors.New("test producer unavailable")

	producer := &handlerTestProducer{
		publishFunc: func(_ context.Context, _ int32, _ walletEvent) error {
			return expectedErr
		},
	}

	e, a, initialState := newHandlerTestEngine(t, producer)

	_, err := e.submit(context.Background(), initialState.WalletGlobalID, handlerTestParams())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected producer error, got %v", err)
	}

	// 明确失败不应触发 Actor 故障终止
	a.mu.Lock()
	failed := a.failed
	a.mu.Unlock()

	if failed {
		t.Fatal("actor should not enter failed state on a definite producer error")
	}

	// Kafka 明确失败时不能提前修改正式余额和版本
	if a.state != initialState {
		t.Fatalf("wallet state changed after failed write: %+v", a.state)
	}

	// 明确失败后 Actor 仍然可以接受下一笔操作
	_, err = e.submit(context.Background(), initialState.WalletGlobalID, handlerTestParams())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected producer error on retry, got %v", err)
	}
}

// TestOperationHandlerReturnsBalanceSnapshot 验证余额查询直接读取运行时状态
func TestOperationHandlerReturnsBalanceSnapshot(t *testing.T) {
	published := make(chan walletEvent, 1)

	producer := &handlerTestProducer{
		publishFunc: func(_ context.Context, _ int32, event walletEvent) error {
			published <- event
			return nil
		},
	}

	e, a, initialState := newHandlerTestEngine(t, producer)

	got, err := e.getBalance(context.Background(), initialState.WalletGlobalID)
	if err != nil {
		t.Fatal(err)
	}

	want := newBalanceSnapshot(initialState)
	if got != want {
		t.Fatalf("unexpected balance snapshot: got %+v, want %+v", got, want)
	}

	a.close()
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 余额查询不能修改钱包状态
	if a.state != initialState {
		t.Fatalf("wallet state changed after query: got %+v, want %+v", a.state, initialState)
	}

	// 余额查询不应产生 Kafka 事件
	select {
	case event := <-published:
		t.Fatalf("unexpected Kafka event: %+v", event)
	default:
	}
}

// TestOperationHandlerSerializesBalanceQuery 验证余额查询与资金操作严格串行
func TestOperationHandlerSerializesBalanceQuery(t *testing.T) {
	publishStarted := make(chan struct{})
	releasePublish := make(chan struct{})

	producer := &handlerTestProducer{
		publishFunc: func(_ context.Context, _ int32, _ walletEvent) error {
			close(publishStarted)
			<-releasePublish
			return nil
		},
	}

	e, a, initialState := newHandlerTestEngine(t, producer)

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(releasePublish) })
	}
	t.Cleanup(releaseAll)

	type submitResponse struct {
		result OperationResult
		err    error
	}

	type balanceResponse struct {
		result BalanceSnapshot
		err    error
	}

	submitDone := make(chan submitResponse, 1)
	balanceDone := make(chan balanceResponse, 1)

	// 先提交资金操作
	go func() {
		result, err := e.submit(
			context.Background(),
			initialState.WalletGlobalID,
			handlerTestParams(),
		)

		submitDone <- submitResponse{
			result: result,
			err:    err,
		}
	}()

	// 等待 Kafka 写入开始
	select {
	case <-publishStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("Kafka publish did not start")
	}

	// Kafka 尚未确认时, 提交余额查询
	go func() {
		result, err := e.getBalance(
			context.Background(),
			initialState.WalletGlobalID,
		)

		balanceDone <- balanceResponse{
			result: result,
			err:    err,
		}
	}()

	// 确认查询已经进入 Mailbox
	// 此时 Actor 仍被前面的资金操作占用
	//nolint:revive
	waitForPendingMailboxMessage(t, a)

	// 资金操作尚未完成, 查询不能提前返回
	select {
	case got := <-balanceDone:
		t.Fatalf("balance query returned before operation finished: %+v", got)
	default:
	}

	// 允许 Kafka 写入成功
	releaseAll()

	// 等待资金操作完成
	var operation submitResponse

	select {
	case operation = <-submitDone:
	case <-time.After(5 * time.Second):
		t.Fatal("wallet operation did not finish")
	}

	if operation.err != nil {
		t.Fatal(operation.err)
	}

	// 等待余额查询完成
	var balance balanceResponse

	select {
	case balance = <-balanceDone:
	case <-time.After(5 * time.Second):
		t.Fatal("balance query did not finish")
	}

	if balance.err != nil {
		t.Fatal(balance.err)
	}

	// 查询必须返回资金操作完成后的余额
	want := newBalanceSnapshot(initialState)
	want.Balance += 100
	want.Version++

	if balance.result != want {
		t.Fatalf("unexpected balance snapshot: got %+v, want %+v", balance.result, want)
	}
}

// TestOperationHandlerRejectsPendingBalanceQueryOnFailure 验证钱包故障后拒绝待处理的余额查询
func TestOperationHandlerRejectsPendingBalanceQueryOnFailure(t *testing.T) {
	publishStarted := make(chan struct{})
	releasePublish := make(chan struct{})

	producer := &handlerTestProducer{
		publishFunc: func(_ context.Context, _ int32, _ walletEvent) error {
			close(publishStarted)
			<-releasePublish

			return fmt.Errorf("%w: test timeout", ErrKafkaWriteUncertain)
		},
	}

	e, a, initialState := newHandlerTestEngine(t, producer)

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(releasePublish) })
	}
	t.Cleanup(releaseAll)

	submitDone := make(chan error, 1)
	balanceDone := make(chan error, 1)

	// 先提交一笔资金操作
	go func() {
		_, err := e.submit(
			context.Background(),
			initialState.WalletGlobalID,
			handlerTestParams(),
		)
		submitDone <- err
	}()

	select {
	case <-publishStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("Kafka publish did not start")
	}

	// 资金操作尚未结束, 提交余额查询
	go func() {
		_, err := e.getBalance(
			context.Background(),
			initialState.WalletGlobalID,
		)
		balanceDone <- err
	}()

	// 确认查询已经进入 Mailbox
	waitForPendingMailboxMessage(t, a)

	// 模拟 Kafka 写入结果不明确
	releaseAll()

	// 当前资金操作应返回原始错误
	select {
	case err := <-submitDone:
		if !errors.Is(err, ErrKafkaWriteUncertain) {
			t.Fatalf("expected ErrKafkaWriteUncertain, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wallet operation did not return")
	}

	// 排队中的余额查询应收到 Actor 故障错误
	select {
	case err := <-balanceDone:
		if !errors.Is(err, ErrWalletFailed) {
			t.Fatalf("expected ErrWalletFailed, got %v", err)
		}
		if !errors.Is(err, ErrKafkaWriteUncertain) {
			t.Fatalf("expected ErrKafkaWriteUncertain, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("pending balance query did not receive rejection")
	}

	// 等待 Actor 完全退出
	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	// Kafka 结果不明确时, 正式钱包状态不能提前修改
	if a.state != initialState {
		t.Fatalf("wallet state changed after uncertain write: got %+v, want %+v", a.state, initialState)
	}
}
