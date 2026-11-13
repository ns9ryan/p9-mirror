// ============================================================================
// 钱包 Actor 测试
//
// 测试 Actor 的创建、消息处理、并发隔离和生命周期
// 重点验证同一个 Actor 串行处理消息, 不同 Actor 可以并发执行
// 同时验证关闭后排空消息、拒绝新消息和等待退出等行为
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
		if err := a.wait(context.Background()); err != nil {
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
		// 测试正常结束和提前退出都可能释放 Handler, 只允许关闭一次
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

// ---- 关闭与等待 ----

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

		// 记录实际处理顺序, 供测试在 Actor 退出后检查
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
	<-started

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
	<-started

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
