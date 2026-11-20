package wallet

import (
	"context"
	"errors"
	"fmt"
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
