// ============================================================================
// 钱包 Actor 测试
//
// 测试 Actor 的创建、消息处理、并发隔离和生命周期
// 重点验证同一个 Actor 串行处理消息, 不同 Actor 可以并发执行
// 同时验证正常关闭排空、故障终止和等待退出等行为
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---- 测试辅助方法 ----

// testWalletState 创建一份有效的钱包状态作为测试基准
func testWalletState() State {
	return State{
		WalletID:       1,
		WalletGlobalID: uuid.New(),
		MemberGlobalID: uuid.New(),
		OperatorCode:   "op001",
		CurrencyCode:   "USDT",
	}
}

// newTestWorkerPool 创建测试用 Worker Pool
func newTestWorkerPool(t *testing.T, workerCount int) *workerPool {
	t.Helper()

	pool, err := newWorkerPool(workerCount)
	if err != nil {
		t.Fatal(err)
	}

	// Actor 的清理会在此后注册, t.Cleanup 按后进先出执行
	// 确保先等待 Actor 退出, 再关闭 Worker Pool
	t.Cleanup(func() {
		if err := pool.close(); err != nil {
			t.Errorf("close worker pool: %v", err)
		}
	})

	return pool
}

// newTestActor 创建绑定测试 Worker Pool 的 Actor
func newTestActor(t *testing.T, pool *workerPool, mailboxCapacity int, handler actorHandler) *actor {
	t.Helper()

	a, err := newActor(pool, testWalletState(), mailboxCapacity, handler)
	if err != nil {
		t.Fatal(err)
	}

	// 测试提前退出时也要关闭并等待 Actor, 避免遗留后台任务
	t.Cleanup(func() {
		a.close()
		if err := a.wait(context.Background()); err != nil && !errors.Is(err, ErrWalletFailed) {
			t.Errorf("wait actor: %v", err)
		}
	})

	return a
}

// waitSignalContext 在 wait 进入 select 时通知测试
type waitSignalContext struct {
	context.Context
	doneCalled chan struct{}
}

// Done 通知测试 wait 已通过 Context 预检查
func (c *waitSignalContext) Done() <-chan struct{} {
	select {
	case c.doneCalled <- struct{}{}:
	default:
	}

	return c.Context.Done()
}

// ---- Actor 创建 ----

// TestNewActorRejectsInvalidConfig 验证 Actor 拒绝无效配置
func TestNewActorRejectsInvalidConfig(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	handler := actorHandler(func(_ *State, _ any) {})

	tests := []struct {
		name            string
		pool            *workerPool
		state           State
		mailboxCapacity int
		handler         actorHandler
	}{
		{
			name:            "nil worker pool",
			pool:            nil,
			state:           testWalletState(),
			mailboxCapacity: 1,
			handler:         handler,
		},
		{
			name:            "invalid wallet state",
			pool:            pool,
			state:           State{},
			mailboxCapacity: 1,
			handler:         handler,
		},
		{
			name:            "invalid mailbox capacity",
			pool:            pool,
			state:           testWalletState(),
			mailboxCapacity: 0,
			handler:         handler,
		},
		{
			name:            "nil handler",
			pool:            pool,
			state:           testWalletState(),
			mailboxCapacity: 1,
			handler:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := newActor(tt.pool, tt.state, tt.mailboxCapacity, tt.handler); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// ---- 消息处理与并发 ----

// TestActorProcessesMessagesInOrder 验证同一个 Actor 按 FIFO 顺序串行处理消息
func TestActorProcessesMessagesInOrder(t *testing.T) {
	// 使用多个 Worker, 验证同一个 Actor 不会因为 Worker 数量增加而并发处理
	pool := newTestWorkerPool(t, 3)

	got := make([]int, 0, 3)
	a := newTestActor(t, pool, 3, func(state *State, message any) {
		got = append(got, message.(int))
		state.Version++
	})

	// 连续提交三条消息, 邮箱容量足够容纳全部消息
	for _, message := range []int{1, 2, 3} {
		if err := a.enqueue(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}

	// 关闭后不再接收新消息, 但已经入队的消息仍需处理完毕
	a.close()
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	// wait 返回后 Handler 已全部执行结束, 可以安全检查处理结果
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("expected [1 2 3], got %v", got)
	}

	if a.state.Version != 3 {
		t.Fatalf("expected state version 3, got %d", a.state.Version)
	}
}

// TestActorRunsDifferentWalletsConcurrently 验证不同 Actor 可以并行处理消息
func TestActorRunsDifferentWalletsConcurrently(t *testing.T) {
	pool := newTestWorkerPool(t, 2)

	started := make(chan struct{}, 2)
	release := make(chan struct{})

	handler := func(_ *State, _ any) {
		// 通知测试当前 Handler 已经开始执行
		started <- struct{}{}

		// 暂停 Handler, 让测试确认另一个 Actor 也能同时执行
		<-release
	}

	a1 := newTestActor(t, pool, 1, handler)
	a2 := newTestActor(t, pool, 1, handler)

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := a1.enqueue(context.Background(), "first"); err != nil {
		t.Fatal(err)
	}

	if err := a2.enqueue(context.Background(), "second"); err != nil {
		t.Fatal(err)
	}

	// 必须等两个 Handler 都开始执行, 才能证明不同 Actor 可以并发运行
	for range 2 {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("different actors did not run concurrently")
		}
	}

	releaseAll()
	a1.close()
	a2.close()

	if err := a1.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := a2.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestActorCloseDrainsPendingMessages 验证关闭后仍会处理已经入队的消息
func TestActorCloseDrainsPendingMessages(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan int, 3)

	a := newTestActor(t, pool, 2, func(_ *State, message any) {
		value := message.(int)
		if value == 1 {
			// 第一条消息开始处理后暂停, 为后续消息留出入队时间
			close(started)
			<-release
		}

		processed <- value
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := a.enqueue(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	// 等待第一条消息进入 Handler, 确保后两条消息留在 Mailbox 中
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("actor handler did not start")
	}

	for _, message := range []int{2, 3} {
		if err := a.enqueue(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}

	// 此时 Mailbox 中仍有两条消息, 关闭后必须继续排空
	a.close()
	releaseAll()

	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	got := []int{<-processed, <-processed, <-processed}
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("expected [1 2 3], got %v", got)
	}
}

// TestActorMailboxFull 验证 Actor 邮箱已满时拒绝新消息
func TestActorMailboxFull(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})

	a := newTestActor(t, pool, 1, func(_ *State, message any) {
		if message.(int) == 1 {
			// 第一条消息正在处理, 不占用 Mailbox 的缓冲容量
			close(started)
			<-release
		}
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := a.enqueue(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	// 确保第一条消息已经被 Worker 取出, Handler 正在执行
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("actor handler did not start")
	}

	// 第一条消息正在处理, 第二条消息占满 Mailbox 的唯一缓冲位置
	if err := a.enqueue(context.Background(), 2); err != nil {
		t.Fatal(err)
	}

	// 缓冲区已满, 第三条消息必须被拒绝
	if err := a.enqueue(context.Background(), 3); !errors.Is(err, ErrMailboxFull) {
		t.Fatalf("expected ErrMailboxFull, got %v", err)
	}

	a.close()
	releaseAll()

	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestActorRejectsEnqueueAfterClose 验证 Actor 关闭后拒绝新消息
func TestActorRejectsEnqueueAfterClose(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	a := newTestActor(t, pool, 1, func(_ *State, _ any) {})

	// 空闲 Actor 关闭后应立即退出, 重复关闭不能产生副作用
	a.close()
	a.close()

	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := a.enqueue(context.Background(), "message"); !errors.Is(err, ErrActorClosed) {
		t.Fatalf("expected ErrActorClosed, got %v", err)
	}
}

// ---- 故障终止 ----

// TestActorFailStopsProcessing 验证故障后 Actor 不再处理后续消息
func TestActorFailStopsProcessing(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan int, 3)

	var a *actor
	a = newTestActor(t, pool, 2, func(_ *State, message any) {
		value := message.(int)
		processed <- value

		if value == 1 {
			// 第一条消息开始执行后暂停, 等待测试提交后续消息
			close(started)
			<-release

			// 模拟 Kafka 写入结果不明确
			a.fail(errors.New("kafka write result uncertain"))
		}
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := a.enqueue(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("first actor message did not start")
	}

	// 第一条消息正在执行, 后续消息进入 Mailbox
	for _, message := range []int{2, 3} {
		if err := a.enqueue(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}

	// 允许第一条消息触发故障终止
	releaseAll()

	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	// 只有触发故障的第一条消息允许执行
	select {
	case got := <-processed:
		if got != 1 {
			t.Fatalf("expected only message 1 to be processed, got %d", got)
		}
	default:
		t.Fatal("first message was not processed")
	}

	if len(processed) != 0 {
		t.Fatalf("unexpected pending messages processed: %d", len(processed))
	}
}

// TestActorFailRejectsPendingMessages 验证故障后已排队消息全部收到拒绝结果
func TestActorFailRejectsPendingMessages(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})

	var a *actor
	a = newTestActor(t, pool, 2, func(_ *State, message any) {
		if message == "trigger" {
			close(started)
			<-release

			// 当前消息触发故障, Mailbox 中的消息应全部被拒绝
			a.fail(errors.New("kafka write result uncertain"))
		}
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := a.enqueue(context.Background(), "trigger"); err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("actor handler did not start")
	}

	// 构造两条待处理的资金操作消息
	messages := []operationMessage{
		{reply: make(chan operationReply, 1)},
		{reply: make(chan operationReply, 1)},
	}

	for i := range messages {
		if err := a.enqueue(context.Background(), messages[i]); err != nil {
			t.Fatal(err)
		}
	}

	releaseAll()

	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	// 每条排队消息都必须收到故障结果, 不能让 RPC 一直等待
	for i, message := range messages {
		select {
		case reply := <-message.reply:
			if !errors.Is(reply.err, ErrWalletFailed) {
				t.Errorf("message %d: expected ErrWalletFailed, got %v", i, reply.err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("message %d did not receive rejection", i)
		}
	}
}

// TestActorFailRejectsScheduledMessages 验证等待调度时故障能够清理消息并退出
func TestActorFailRejectsScheduledMessages(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})

	// 占用唯一 Worker, 使目标 Actor 只能等待调度
	blocker := newTestActor(t, pool, 1, func(_ *State, _ any) {
		close(started)
		<-release
	})

	handled := make(chan struct{}, 1)
	target := newTestActor(t, pool, 1, func(_ *State, _ any) {
		handled <- struct{}{}
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := blocker.enqueue(context.Background(), "block"); err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("blocker did not start")
	}

	// 目标 Actor 已进入 Ready Queue, 但 Worker 仍被 blocker 占用
	message := operationMessage{
		reply: make(chan operationReply, 1),
	}

	if err := target.enqueue(context.Background(), message); err != nil {
		t.Fatal(err)
	}

	cause := errors.New("kafka write result uncertain")
	event := walletEvent{
		OperationID:    uuid.New(),
		WalletGlobalID: target.state.WalletGlobalID,
	}

	target.failOperation(event, cause)

	// 排队消息必须立即收到拒绝结果
	select {
	case reply := <-message.reply:
		if !errors.Is(reply.err, ErrWalletFailed) || !errors.Is(reply.err, cause) {
			t.Fatalf("unexpected rejection error: %v", reply.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pending message did not receive rejection")
	}

	// Worker 尚未处理 Ready Queue 中的目标 Actor, 此时不能提前退出
	select {
	case <-target.done:
		t.Fatal("target actor exited before scheduled task was completed")
	default:
	}

	// 释放 Worker, 让目标 Actor 完成最终退出
	releaseAll()

	if err := target.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	if len(handled) != 0 {
		t.Fatal("failed actor unexpectedly processed pending message")
	}

	blocker.close()
	if err := blocker.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestActorFailRejectsNewMessages 验证故障后拒绝新消息
func TestActorFailRejectsNewMessages(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	a := newTestActor(t, pool, 1, func(_ *State, _ any) {})

	a.fail(errors.New("kafka write result uncertain"))

	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	if err := a.enqueue(context.Background(), "message"); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed on enqueue, got %v", err)
	}
}

// TestActorFailIsIdempotent 验证重复触发故障不会覆盖首次故障原因
func TestActorFailIsIdempotent(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	a := newTestActor(t, pool, 1, func(_ *State, _ any) {})

	firstErr := errors.New("first failure")
	secondErr := errors.New("second failure")

	a.fail(firstErr)
	a.fail(secondErr)

	err := a.wait(context.Background())
	if !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}

	if !errors.Is(err, firstErr) {
		t.Fatalf("expected first failure to be preserved, got %v", err)
	}

	if errors.Is(err, secondErr) {
		t.Fatalf("unexpected second failure in actor error: %v", err)
	}

	// 重复故障不能重复关闭 done
	select {
	case <-a.done:
	default:
		t.Fatal("actor did not stop after failure")
	}
}

// TestActorFailPreservesOperationEvent 验证故障后保留原始资金事件
func TestActorFailPreservesOperationEvent(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	a := newTestActor(t, pool, 1, func(_ *State, _ any) {})

	operationID := uuid.New()
	cause := errors.New("kafka write result uncertain")

	event := walletEvent{
		OperationID:    operationID,
		WalletGlobalID: a.state.WalletGlobalID,
		Ledgers: []walletLedgerEvent{
			{
				LedgerGlobalID: uuid.New(),
				ChangeAmount:   100,
			},
		},
	}

	a.failOperation(event, cause)

	// 修改原事件, 不应影响 Actor 保存的故障信息
	event.Ledgers[0].ChangeAmount = 200

	failure, ok := a.failureInfo()
	if !ok {
		t.Fatal("expected actor failure info")
	}

	if failure.event.OperationID != operationID {
		t.Fatalf("unexpected operation ID: %s", failure.event.OperationID)
	}

	if failure.event.Ledgers[0].ChangeAmount != 100 {
		t.Fatal("stored failure event was unexpectedly modified")
	}

	if !errors.Is(failure.cause, cause) {
		t.Fatalf("expected original failure cause, got %v", failure.cause)
	}

	// 修改读取结果, 也不应影响 Actor 内部保存的数据
	failure.event.Ledgers[0].ChangeAmount = 300

	stored, ok := a.failureInfo()
	if !ok || stored.event.Ledgers[0].ChangeAmount != 100 {
		t.Fatal("failure info was modified through returned snapshot")
	}

	if err := a.wait(context.Background()); !errors.Is(err, ErrWalletFailed) ||
		!errors.Is(err, cause) {
		t.Fatalf("unexpected actor error: %v", err)
	}
}

// ---- 关闭与等待 ----

// TestActorWaitContextCanceled 验证等待 Actor 退出支持 Context 取消
func TestActorWaitContextCanceled(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	a := newTestActor(t, pool, 1, func(_ *State, _ any) {})

	// 预先取消 Context, 验证 wait 不会一直等待 Actor 退出
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := a.wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// Context 取消只结束本次等待, 不会关闭 Actor
	a.close()
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestActorWaitContextCanceledWhileWaiting 验证等待期间取消 Context
func TestActorWaitContextCanceledWhileWaiting(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})

	a := newTestActor(t, pool, 1, func(_ *State, _ any) {
		// 暂停 Handler, 确保 Actor 不会提前退出
		close(started)
		<-release
	})

	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseAll)

	if err := a.enqueue(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}

	// 等待 Handler 开始执行, 确保 Actor 仍处于运行状态
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("actor handler did not start")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	waitCtx := &waitSignalContext{
		Context:    ctx,
		doneCalled: make(chan struct{}, 1),
	}

	waitResult := make(chan error, 1)

	// 在独立 Goroutine 中等待 Actor 退出
	go func() {
		waitResult <- a.wait(waitCtx)
	}()

	// 确保 wait 已通过预检查并调用 Done(), 即将进入 select 等待
	select {
	case <-waitCtx.doneCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("actor wait did not enter select")
	}

	// 等待期间取消 Context, 本次等待应立即返回
	cancel()

	select {
	case err := <-waitResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("actor wait did not return after context cancellation")
	}

	// Context 取消只结束本次等待, 不应该关闭 Actor
	select {
	case <-a.done:
		t.Fatal("actor exited after wait context cancellation")
	default:
	}

	// 释放 Handler, 再正常关闭并等待 Actor 退出
	a.close()
	releaseAll()

	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
