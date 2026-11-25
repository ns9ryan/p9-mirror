// ============================================================================
// 钱包 Engine 测试
//
// 测试 Engine 的创建、路由、Actor 管理、并发访问和整体关闭
// 同时验证资金操作提交、结果返回和 Context 取消行为
// ============================================================================

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

// newTestEngine 创建测试 Engine, 并在测试结束时关闭和等待全部资源
// expectedWaitErr 用于声明故障测试中允许出现的预期退出错误
func newTestEngine(t *testing.T, shardCount, workerCount int, handler actorHandler, expectedWaitErr ...error) *engine {
	t.Helper()

	e, err := newEngine(shardCount, workerCount, 2, handler)
	if err != nil {
		t.Fatal(err)
	}

	// 先关闭全部 Shard, 再等待 Actor 和 Worker Pool 退出
	t.Cleanup(func() {
		e.close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.wait(ctx); err != nil && !isOnlyExpectedError(err, expectedWaitErr) {
			t.Errorf("wait engine: %v", err)
		}
	})

	return e
}

// isOnlyExpectedError 检查错误树中的错误是否都属于明确允许的预期错误
func isOnlyExpectedError(err error, expected []error) bool {
	if err == nil {
		return true
	}

	// 先匹配完整错误节点, 保留 errors.Join 子树的整体语义
	for _, target := range expected {
		if target != nil && err == target {
			return true
		}
	}

	// errors.Join 可能包含多个独立错误, 必须逐个检查
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		errs := joined.Unwrap()
		if len(errs) == 0 {
			return false
		}

		for _, item := range errs {
			if !isOnlyExpectedError(item, expected) {
				return false
			}
		}
		return true
	}

	// 普通错误包装只有一个底层错误, 继续检查底层节点
	if wrapped := errors.Unwrap(err); wrapped != nil {
		return isOnlyExpectedError(wrapped, expected)
	}

	return false
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
//  4. 释放 Handler, 重新等待 Actor 正常退出
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

// ---- 余额查询 ----

// TestEngineGetBalanceReturnsSnapshot 验证 Engine 能够返回钱包余额快照
func TestEngineGetBalanceReturnsSnapshot(t *testing.T) {
	e := newTestEngine(t, 8, 4, func(state *State, message any) {
		m, ok := message.(balanceQueryMessage)
		if !ok || m.reply == nil {
			return
		}

		m.reply <- balanceQueryReply{
			result: newBalanceSnapshot(*state),
		}
	})

	state := testWalletState()
	state.Balance = 500
	state.FrozenBalance = 20
	state.UnsettledBalance = 30
	state.Version = 7

	a, err := e.getOrCreate(state)
	if err != nil {
		t.Fatal(err)
	}

	got, err := e.getBalance(context.Background(), state.WalletGlobalID)
	if err != nil {
		t.Fatal(err)
	}

	want := newBalanceSnapshot(state)
	if got != want {
		t.Fatalf("unexpected balance snapshot: got %+v, want %+v", got, want)
	}

	// 等待 Actor 退出后, 确认查询没有修改钱包状态
	a.close()
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	if a.state != state {
		t.Fatalf("wallet state changed after balance query: got %+v, want %+v", a.state, state)
	}
}

// TestEngineGetBalanceRejectsInvalidRequests 验证 Engine 拒绝无效余额查询请求
func TestEngineGetBalanceRejectsInvalidRequests(t *testing.T) {
	var nilEngine *engine

	if _, err := nilEngine.getBalance(context.Background(), uuid.New()); !errors.Is(err, errNilEngine) {
		t.Fatalf("expected errNilEngine, got %v", err)
	}

	e := newTestEngine(t, 2, 1, func(_ *State, _ any) {})

	// Context 不能为空
	if _, err := e.getBalance(nil, uuid.New()); !errors.Is(err, errNilEngineContext) {
		t.Fatalf("expected errNilEngineContext, got %v", err)
	}

	// 已取消的 Context 不应进入查询流程
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := e.getBalance(ctx, uuid.New()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// 钱包全局标识不能为空
	if _, err := e.getBalance(context.Background(), uuid.Nil); !errors.Is(err, ErrInvalidWalletGlobalID) {
		t.Fatalf("expected ErrInvalidWalletGlobalID, got %v", err)
	}

	// 尚未初始化的 Actor 不允许查询
	if _, err := e.getBalance(context.Background(), uuid.New()); !errors.Is(err, ErrWalletActorNotFound) {
		t.Fatalf("expected ErrWalletActorNotFound, got %v", err)
	}

	// Engine 关闭后拒绝查询
	e.close()

	if _, err := e.getBalance(context.Background(), uuid.New()); !errors.Is(err, ErrEngineClosed) {
		t.Fatalf("expected ErrEngineClosed, got %v", err)
	}
}

// TestEngineGetBalanceContextCanceledAfterEnqueue 验证取消等待不会阻止已入队的查询消息
func TestEngineGetBalanceContextCanceledAfterEnqueue(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan struct{})

	e := newTestEngine(t, 2, 1, func(state *State, message any) {
		m, ok := message.(balanceQueryMessage)
		if !ok || m.reply == nil {
			return
		}

		close(started)
		<-release

		m.reply <- balanceQueryReply{
			result: newBalanceSnapshot(*state),
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

	result := make(chan error, 1)

	go func() {
		_, err := e.getBalance(ctx, state.WalletGlobalID)
		result <- err
	}()

	// 确保查询已经进入 Actor Handler
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("balance query handler did not start")
	}

	// 取消当前调用方的等待
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("balance query did not return after context cancellation")
	}

	// Actor 仍应正常完成已经接收的查询
	releaseAll()

	select {
	case <-processed:
	case <-time.After(5 * time.Second):
		t.Fatal("accepted balance query did not finish")
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

// TestEngineSubmitRoutesAndReturnsResult 验证资金操作路由、Actor 绑定及结果返回
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

	// 保存 Actor 指针, 后续验证 Engine 是否正确绑定消息
	a, err := e.getOrCreate(state)
	if err != nil {
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
	if message.actor != a {
		t.Fatal("operation message is not bound to its wallet actor")
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

// TestEngineSubmitRejectsFailedWallet 验证钱包故障后 Engine 拒绝后续资金操作
func TestEngineSubmitRejectsFailedWallet(t *testing.T) {
	handled := make(chan struct{}, 2)

	kafkaTimeoutErr := errors.New("test kafka timeout")

	uncertainErr := errors.Join(ErrKafkaWriteUncertain, kafkaTimeoutErr)

	e := newTestEngine(t, 2, 1, func(state *State, message any) {
		m := message.(operationMessage)
		handled <- struct{}{}

		// 先计算候选结果, 构造完整的原始资金事件
		prepared, err := prepareRecord(*state, m.params)
		if err != nil {
			m.reply <- operationReply{err: err}
			return
		}

		event := newWalletEvent(*state, m.params, prepared)

		// 模拟 Kafka 写入结果不明确, 保留原始事件并触发钱包故障
		m.actor.failOperation(event, uncertainErr)

		// 当前操作仍需将原始错误返回给 RPC 调用方
		m.reply <- operationReply{err: uncertainErr}
	}, ErrWalletFailed, ErrKafkaWriteUncertain, kafkaTimeoutErr)

	state := testWalletState()
	a, err := e.getOrCreate(state)
	if err != nil {
		t.Fatal(err)
	}

	params := testSubmitParams()

	// 第一次操作触发钱包故障, 调用方收到原始错误
	_, err = e.submit(context.Background(), state.WalletGlobalID, params)
	if !errors.Is(err, ErrKafkaWriteUncertain) {
		t.Fatalf("expected ErrKafkaWriteUncertain, got %v", err)
	}

	// 等待 Actor 完全退出, 确保故障状态已经生效
	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	// Kafka 未确认时, 正式钱包状态不能发生变化
	if a.state != state {
		t.Fatalf("wallet state changed before Kafka confirmation: got %+v, want %+v", a.state, state)
	}

	// 故障信息必须保留原始资金事件及操作标识
	failure, ok := a.failureInfo()
	if !ok {
		t.Fatal("expected wallet failure info")
	}

	if failure.event.OperationID != params.OperationID {
		t.Fatalf("unexpected operation ID: %s", failure.event.OperationID)
	}

	if failure.event.WalletGlobalID != state.WalletGlobalID {
		t.Fatal("failure event wallet ID mismatch")
	}

	if !errors.Is(failure.cause, ErrKafkaWriteUncertain) {
		t.Fatalf("unexpected failure cause: %v", failure.cause)
	}

	// 故障 Actor 必须继续保留在 Shard 中, 不能被重新创建
	reused, err := e.getOrCreate(state)
	if err != nil {
		t.Fatal(err)
	}
	if reused != a {
		t.Fatal("failed wallet actor was unexpectedly replaced")
	}

	// 同一个钱包再次提交操作, Engine 必须返回故障错误
	_, err = e.submit(context.Background(), state.WalletGlobalID, testSubmitParams())
	if !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	// Handler 只能执行第一次, 后续操作不能进入处理流程
	if got := len(handled); got != 1 {
		t.Fatalf("expected handler to execute once, got %d", got)
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

// TestEngineSubmitContextCanceledBeforeEnqueue 验证入队前取消不会被标记为已接收
func TestEngineSubmitContextCanceledBeforeEnqueue(t *testing.T) {
	e := newTestEngine(t, 2, 1, func(_ *State, _ any) {})

	state := testWalletState()
	if _, err := e.getOrCreate(state); err != nil {
		t.Fatal(err)
	}

	// 提交前取消 Context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := e.submit(ctx, state.WalletGlobalID, testSubmitParams())

	// 入队前取消应返回原始 Context 错误
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// 尚未进入 Mailbox, 不应标记为操作已被接收
	if errors.Is(err, ErrOperationWaitInterrupted) {
		t.Fatalf("unexpected ErrOperationWaitInterrupted: %v", err)
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
		// 必须同时匹配两个错误
		if !errors.Is(err, ErrOperationWaitInterrupted) {
			t.Fatalf("expected ErrOperationWaitInterrupted, got %v", err)
		}
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
