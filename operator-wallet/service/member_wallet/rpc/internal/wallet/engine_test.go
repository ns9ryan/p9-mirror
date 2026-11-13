package wallet

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// ---- 测试辅助方法 ----

// newTestEngine 创建测试 Engine, 并在测试结束时完成整体关闭
func newTestEngine(t *testing.T, shardCount, workerCount int, handler actorHandler) *engine {
	t.Helper()

	e, err := newEngine(shardCount, workerCount, 2, handler)
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup 先关闭全部 Shard, 再等待 Actor 和 Worker Pool 退出
	t.Cleanup(func() {
		e.close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.wait(ctx); err != nil {
			t.Errorf("wait engine: %v", err)
		}
	})

	return e
}

// ---- 创建与路由 ----

// TestNewEngineRejectsInvalidConfig 验证 Engine 拒绝无效配置
func TestNewEngineRejectsInvalidConfig(t *testing.T) {
	handler := actorHandler(func(_ *State, _ any) {})

	tests := []struct {
		name            string       // 测试场景
		shardCount      int          // Shard 数量
		workerCount     int          // Worker 数量
		mailboxCapacity int          // Actor Mailbox 容量
		handler         actorHandler // 钱包消息处理函数
	}{
		{"invalid shard count", 0, 1, 1, handler},
		{"invalid worker count", 1, 0, 1, handler},
		{"invalid mailbox capacity", 1, 1, 0, handler},
		{"nil handler", 1, 1, 1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newEngine(tt.shardCount, tt.workerCount, tt.mailboxCapacity, tt.handler)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestEngineRoutesWalletsToShards 验证 Engine 将钱包交给正确的 Shard
func TestEngineRoutesWalletsToShards(t *testing.T) {
	e := newTestEngine(t, 8, 4, func(_ *State, _ any) {})

	for i := range 32 {
		state := testWalletState()
		state.WalletID = int64(i + 1)

		// 先根据 Router 确定目标 Shard
		index, err := e.router.route(state.WalletGlobalID)
		if err != nil {
			t.Fatal(err)
		}

		// Engine 应将 Actor 创建在目标 Shard 中
		a, err := e.getOrCreate(state)
		if err != nil {
			t.Fatal(err)
		}

		found, err := e.shards[index].get(state.WalletGlobalID)
		if err != nil {
			t.Fatal(err)
		}
		if found != a {
			t.Fatalf("wallet routed to wrong shard: expected %d", index)
		}

		// 通过 Engine 查询时也必须返回同一个 Actor
		found, err = e.get(state.WalletGlobalID)
		if err != nil || found != a {
			t.Fatalf("engine returned unexpected actor: actor=%v, err=%v", found, err)
		}

		// 其他 Shard 不应该登记该钱包
		for j, s := range e.shards {
			if j == index {
				continue
			}

			other, err := s.get(state.WalletGlobalID)
			if err != nil || other != nil {
				t.Fatalf("wallet unexpectedly found in shard %d", j)
			}
		}
	}
}

// TestEngineReusesWalletActor 验证相同钱包复用 Actor, 且拒绝身份冲突
func TestEngineReusesWalletActor(t *testing.T) {
	e := newTestEngine(t, 4, 2, func(_ *State, _ any) {})
	state := testWalletState()

	// 未初始化的钱包应返回 nil, nil
	missing, err := e.get(state.WalletGlobalID)
	if err != nil || missing != nil {
		t.Fatalf("expected missing actor, got actor=%v, err=%v", missing, err)
	}

	first, err := e.getOrCreate(state)
	if err != nil {
		t.Fatal(err)
	}

	// 重复初始化不会覆盖 Actor 已有的运行时状态
	updated := state
	updated.Balance = 100
	updated.Version = 2

	second, err := e.getOrCreate(updated)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("same wallet created multiple actors")
	}
	if first.state.Balance != 0 || first.state.Version != 0 {
		t.Fatal("existing wallet state was overwritten")
	}

	// 同一个钱包标识对应不同厅时, Shard 应拒绝复用
	invalid := state
	invalid.OperatorCode = "op002"

	if _, err := e.getOrCreate(invalid); !errors.Is(err, ErrShardWalletIdentityMismatch) {
		t.Fatalf("expected ErrShardWalletIdentityMismatch, got %v", err)
	}
}

// TestEngineConcurrentGetOrCreate 验证 Engine 并发访问不会重复创建 Actor
//
// 测试流程:
//  1. 启动 64 个 Goroutine 等待统一信号
//  2. 所有请求同时调用 Engine.getOrCreate
//  3. 检查每个请求返回的 Actor 指针一致
//  4. 验证目标 Shard 最终只登记一个 Actor
func TestEngineConcurrentGetOrCreate(t *testing.T) {
	e := newTestEngine(t, 8, 4, func(_ *State, _ any) {})
	state := testWalletState()

	const requestCount = 64

	type result struct {
		actor *actor // 返回的钱包 Actor
		err   error  // 创建或获取时的错误
	}

	start := make(chan struct{})
	results := make(chan result, requestCount)

	// 让所有 Goroutine 等待同一个启动信号
	for range requestCount {
		go func() {
			<-start

			a, err := e.getOrCreate(state)
			results <- result{actor: a, err: err}
		}()
	}

	// 同时释放全部请求, 制造并发创建竞争
	close(start)

	var first *actor
	for range requestCount {
		select {
		case got := <-results:
			if got.err != nil {
				t.Fatal(got.err)
			}
			if first == nil {
				first = got.actor
			}
			if got.actor != first {
				t.Fatal("concurrent requests created multiple actors")
			}

		case <-time.After(5 * time.Second):
			t.Fatal("concurrent getOrCreate did not finish")
		}
	}

	// 所有请求必须最终落到同一个 Shard, 并共享唯一 Actor
	index, err := e.router.route(state.WalletGlobalID)
	if err != nil {
		t.Fatal(err)
	}
	if len(e.shards[index].actors) != 1 {
		t.Fatalf("expected one actor, got %d", len(e.shards[index].actors))
	}
}

// ---- 关闭与等待 ----

// TestEngineWaitRequiresClose 验证等待前必须先关闭 Engine
func TestEngineWaitRequiresClose(t *testing.T) {
	e := newTestEngine(t, 2, 1, func(_ *State, _ any) {})

	if err := e.wait(context.Background()); !errors.Is(err, ErrEngineNotClosed) {
		t.Fatalf("expected ErrEngineNotClosed, got %v", err)
	}
}

// TestEngineCloseDrainsPendingMessages 验证 Engine 关闭后仍会处理已接受的消息
//
// 测试流程:
//  1. 创建两个钱包 Actor, 让第一个 Actor 暂停在 Handler 内
//  2. 继续提交消息, 确保关闭时仍有待处理消息
//  3. 关闭 Engine, 验证拒绝新请求
//  4. 恢复 Handler, 等待所有 Actor 和 Worker 完全退出
//  5. 检查关闭前已接受的消息全部处理
func TestEngineCloseDrainsPendingMessages(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan int, 3)

	e := newTestEngine(t, 4, 2, func(_ *State, message any) {
		value := message.(int)

		if value == 1 {
			close(started)
			<-release
		}

		processed <- value
	})

	// 避免测试提前退出时 Handler 一直阻塞
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)

	state1 := testWalletState()
	state2 := testWalletState()
	state2.WalletID = 2

	a1, err := e.getOrCreate(state1)
	if err != nil {
		t.Fatal(err)
	}

	a2, err := e.getOrCreate(state2)
	if err != nil {
		t.Fatal(err)
	}

	// 确认第一条消息正在处理
	if err := a1.enqueue(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first actor handler did not start")
	}

	// 消息 2 留在 Actor1 的 Mailbox, 消息 3 交给 Actor2
	if err := a1.enqueue(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	if err := a2.enqueue(context.Background(), 3); err != nil {
		t.Fatal(err)
	}

	// Engine 关闭后停止新请求, 但仍需排空已接受的消息
	e.close()
	e.close()

	if _, err := e.getOrCreate(state1); !errors.Is(err, ErrEngineClosed) {
		t.Fatalf("expected ErrEngineClosed, got %v", err)
	}
	if _, err := e.get(state1.WalletGlobalID); !errors.Is(err, ErrEngineClosed) {
		t.Fatalf("expected ErrEngineClosed, got %v", err)
	}
	if err := a1.enqueue(context.Background(), 4); !errors.Is(err, ErrActorClosed) {
		t.Fatalf("expected ErrActorClosed, got %v", err)
	}

	// 恢复 Handler, 等待全部 Actor 和 Worker Pool 退出
	releaseAll()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.wait(ctx); err != nil {
		t.Fatal(err)
	}

	// 不同钱包允许并发处理, 这里只校验消息完整性
	seen := make(map[int]bool)
	for range 3 {
		seen[<-processed] = true
	}
	for _, message := range []int{1, 2, 3} {
		if !seen[message] {
			t.Fatalf("message %d was not processed", message)
		}
	}

	// 所有 Actor 退出后, Engine 必须已经关闭共享 Worker Pool
	if err := e.pool.schedule(a1); !errors.Is(err, ErrWorkerPoolClosed) {
		t.Fatalf("expected ErrWorkerPoolClosed, got %v", err)
	}

	// 重复等待已关闭的 Engine 不应产生副作用
	if err := e.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestEngineWaitContextCanceled 验证等待超时后仍能继续完成关闭
//
// 测试流程:
//  1. 让 Actor 暂停在 Handler 内
//  2. 关闭 Engine, 使用短超时 Context 等待
//  3. 验证等待超时不会强制终止 Actor
//  4. 释放 Handler, 重新等待 Engine 正常退出
func TestEngineWaitContextCanceled(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	e := newTestEngine(t, 2, 1, func(_ *State, _ any) {
		close(started)
		<-release
	})

	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)

	a, err := e.getOrCreate(testWalletState())
	if err != nil {
		t.Fatal(err)
	}

	if err := a.enqueue(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("actor handler did not start")
	}

	e.close()

	// Handler 尚未结束, wait 应在 Context 超时时返回
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := e.wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	// 等待超时不应关闭正在处理消息的 Actor
	select {
	case <-a.done:
		t.Fatal("actor exited before handler completed")
	default:
	}

	// 释放 Handler 后, 再次 wait 应完成剩余关闭流程
	releaseAll()

	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer waitCancel()

	if err := e.wait(waitCtx); err != nil {
		t.Fatal(err)
	}
}

// ---- 并发创建与关闭 ----

// TestEngineConcurrentCreateAndClose 验证并发创建与整体关闭不会遗漏 Actor
//
// 测试流程:
//  1. 准备 64 个不同钱包的创建请求
//  2. 同时触发创建请求与 Engine.close
//  3. 验证请求要么成功, 要么返回 ErrEngineClosed
//  4. 等待全部 Actor 和 Worker 退出, 再验证 Engine 已关闭
//
// Engine.mu 保证创建和关闭互斥
// 创建先取得读锁时, 对应 Actor 会被登记并由关闭流程管理
// close 先取得写锁时, 后续创建请求直接返回 ErrEngineClosed
func TestEngineConcurrentCreateAndClose(t *testing.T) {
	e := newTestEngine(t, 8, 4, func(_ *State, _ any) {})

	const requestCount = 64

	start := make(chan struct{})
	results := make(chan error, requestCount)
	closeDone := make(chan struct{})

	// 每个请求使用独立钱包, 让多个 Shard 同时参与竞争
	for i := range requestCount {
		state := testWalletState()
		state.WalletID = int64(i + 1)

		go func() {
			<-start

			_, err := e.getOrCreate(state)
			results <- err
		}()
	}

	// 同时释放创建请求并启动关闭流程
	close(start)
	go func() {
		e.close()
		close(closeDone)
	}()

	// 关闭前的创建允许成功, 关闭后的创建必须被拒绝
	for range requestCount {
		select {
		case err := <-results:
			if err != nil && !errors.Is(err, ErrEngineClosed) {
				t.Fatalf("unexpected create error: %v", err)
			}

		case <-time.After(5 * time.Second):
			t.Fatal("concurrent create did not finish")
		}
	}

	select {
	case <-closeDone:
	case <-time.After(5 * time.Second):
		t.Fatal("engine close did not finish")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.wait(ctx); err != nil {
		t.Fatal(err)
	}

	// 所有创建和关闭操作结束后, Engine 必须拒绝新请求
	if _, err := e.get(uuid.New()); !errors.Is(err, ErrEngineClosed) {
		t.Fatalf("expected ErrEngineClosed, got %v", err)
	}
}

// ---- 资金操作提交 ----

// testSubmitParams 创建有效的测试资金操作参数
func testSubmitParams() RecordParams {
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

// TestEngineSubmitRoutesAndReturnsResult 验证资金操作路由及结果返回
func TestEngineSubmitRoutesAndReturnsResult(t *testing.T) {
	handled := make(chan operationMessage, 1)

	// 使用模拟 Handler 验证 Engine 提交流程, 不涉及真实 Kafka
	e := newTestEngine(t, 8, 4, func(state *State, message any) {
		m := message.(operationMessage)
		handled <- m

		state.Balance += m.params.Changes[0].ChangeAmount
		state.Version++

		m.reply <- operationReply{
			result: OperationResult{
				OperationID:      m.params.OperationID,
				WalletGlobalID:   state.WalletGlobalID,
				Version:          state.Version,
				Balance:          state.Balance,
				FrozenBalance:    state.FrozenBalance,
				UnsettledBalance: state.UnsettledBalance,
			},
		}
	})

	state := testWalletState()
	state.Balance = 500
	state.Version = 7

	if _, err := e.getOrCreate(state); err != nil {
		t.Fatal(err)
	}

	params := testSubmitParams()

	result, err := e.submit(context.Background(), state.WalletGlobalID, params)
	if err != nil {
		t.Fatal(err)
	}

	// 验证 Actor 返回的操作结果
	if result.OperationID != params.OperationID ||
		result.WalletGlobalID != state.WalletGlobalID ||
		result.Version != 8 ||
		result.Balance != 600 {
		t.Fatalf("unexpected operation result: %+v", result)
	}

	// 验证消息使用了正确的 Kafka Partition
	index, err := e.router.route(state.WalletGlobalID)
	if err != nil {
		t.Fatal(err)
	}

	message := <-handled
	if message.partition != int32(index) {
		t.Fatalf("expected partition %d, got %d", index, message.partition)
	}
	if message.params.OperationID != params.OperationID {
		t.Fatal("operation ID changed during submission")
	}
}

// TestEngineSubmitRejectsInvalidRequests 验证提交入口的基础校验
func TestEngineSubmitRejectsInvalidRequests(t *testing.T) {
	e := newTestEngine(t, 2, 1, func(_ *State, _ any) {})

	walletGlobalID := uuid.New()
	params := testSubmitParams()

	// Context 不能为空
	if _, err := e.submit(nil, walletGlobalID, params); !errors.Is(err, errNilEngineContext) {
		t.Fatalf("expected errNilEngineContext, got %v", err)
	}

	// 已取消的请求不应进入 Engine
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := e.submit(ctx, walletGlobalID, params); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// 无效的记账参数应提前拒绝
	invalid := params
	invalid.Changes = nil

	if _, err := e.submit(context.Background(), walletGlobalID, invalid); !errors.Is(err, ErrInvalidRecordParams) {
		t.Fatalf("expected ErrInvalidRecordParams, got %v", err)
	}

	// 钱包标识不能为空
	if _, err := e.submit(context.Background(), uuid.Nil, params); !errors.Is(err, ErrInvalidWalletGlobalID) {
		t.Fatalf("expected ErrInvalidWalletGlobalID, got %v", err)
	}

	// 尚未初始化的 Actor 不允许直接处理资金操作
	if _, err := e.submit(context.Background(), walletGlobalID, params); !errors.Is(err, ErrWalletActorNotFound) {
		t.Fatalf("expected ErrWalletActorNotFound, got %v", err)
	}

	// Engine 关闭后拒绝新操作
	e.close()

	if _, err := e.submit(context.Background(), walletGlobalID, params); !errors.Is(err, ErrEngineClosed) {
		t.Fatalf("expected ErrEngineClosed, got %v", err)
	}
}

// TestEngineSubmitReturnsHandlerError 验证 Actor 错误能够返回给调用方
func TestEngineSubmitReturnsHandlerError(t *testing.T) {
	expectedErr := errors.New("test producer failure")

	e := newTestEngine(t, 1, 1, func(_ *State, message any) {
		m := message.(operationMessage)
		m.reply <- operationReply{err: expectedErr}
	})

	state := testWalletState()
	if _, err := e.getOrCreate(state); err != nil {
		t.Fatal(err)
	}

	_, err := e.submit(context.Background(), state.WalletGlobalID, testSubmitParams())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected producer failure, got %v", err)
	}
}

// TestEngineSubmitDoesNotBlockClose 验证等待 Actor 结果时不持有 Engine 读锁
func TestEngineSubmitDoesNotBlockClose(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	e := newTestEngine(t, 2, 1, func(state *State, message any) {
		m := message.(operationMessage)

		close(started)
		<-release

		m.reply <- operationReply{
			result: OperationResult{
				OperationID:    m.params.OperationID,
				WalletGlobalID: state.WalletGlobalID,
			},
		}
	})

	// 测试提前退出时也必须释放 Handler
	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	state := testWalletState()
	if _, err := e.getOrCreate(state); err != nil {
		t.Fatal(err)
	}

	submitDone := make(chan error, 1)

	go func() {
		_, err := e.submit(context.Background(), state.WalletGlobalID, testSubmitParams())
		submitDone <- err
	}()

	// 确认消息已经进入 Actor, 且 Handler 尚未完成
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("actor handler did not start")
	}

	// Engine.close 不应被 submit 的结果等待阻塞
	closeDone := make(chan struct{})

	go func() {
		e.close()
		close(closeDone)
	}()

	select {
	case <-closeDone:
	case <-time.After(5 * time.Second):
		t.Fatal("engine close was blocked by submit")
	}

	// 关闭 Engine 不应取消已接受的资金操作
	select {
	case err := <-submitDone:
		t.Fatalf("submit returned before handler completed: %v", err)
	default:
	}

	releaseAll()

	select {
	case err := <-submitDone:
		if err != nil {
			t.Fatalf("submit failed after engine close: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("submit did not finish")
	}
}

// TestEngineSubmitContextCanceledAfterEnqueue 验证取消等待不会取消已入队消息
func TestEngineSubmitContextCanceledAfterEnqueue(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan struct{})

	e := newTestEngine(t, 2, 1, func(state *State, message any) {
		m := message.(operationMessage)

		close(started)
		<-release

		// 模拟资金操作已完成, 即使 RPC 不再等待也应能通知结果
		m.reply <- operationReply{
			result: OperationResult{
				OperationID:    m.params.OperationID,
				WalletGlobalID: state.WalletGlobalID,
			},
		}

		close(processed)
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	state := testWalletState()
	if _, err := e.getOrCreate(state); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	submitDone := make(chan error, 1)

	go func() {
		_, err := e.submit(ctx, state.WalletGlobalID, testSubmitParams())
		submitDone <- err
	}()

	// 等待 Actor 开始处理, 确保消息已经入队
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("actor handler did not start")
	}

	// 取消的是 RPC 等待, 不是已经接受的资金操作
	cancel()

	select {
	case err := <-submitDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("submit did not respond to cancellation")
	}

	// 恢复 Handler, 确认操作仍能处理完毕
	releaseAll()

	select {
	case <-processed:
	case <-time.After(5 * time.Second):
		t.Fatal("accepted operation did not finish")
	}
}
