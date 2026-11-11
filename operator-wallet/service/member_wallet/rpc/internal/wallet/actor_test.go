// ============================================================================
// 钱包 Actor 测试
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

// testWalletState 创建测试用的钱包状态
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

	t.Cleanup(func() {
		if err := pool.close(); err != nil {
			t.Errorf("close worker pool: %v", err)
		}
	})

	return pool
}

// newTestActor 创建测试用 Actor
func newTestActor(t *testing.T, pool *workerPool, mailboxCapacity int, handler actorHandler) *actor {
	t.Helper()

	a, err := newActor(pool, testWalletState(), mailboxCapacity, handler)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		a.close()
		if err := a.wait(context.Background()); err != nil {
			t.Errorf("wait actor: %v", err)
		}
	})

	return a
}

// TestNewActorRejectsInvalidConfig 验证 Actor 拒绝无效配置
func TestNewActorRejectsInvalidConfig(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	handler := actorHandler(func(_ *State, _ any) {})

	tests := []struct {
		name    string
		pool    *workerPool
		state   State
		size    int
		handler actorHandler
	}{
		{
			name:    "nil worker pool",
			pool:    nil,
			state:   testWalletState(),
			size:    1,
			handler: handler,
		},
		{
			name:    "invalid wallet state",
			pool:    pool,
			state:   State{},
			size:    1,
			handler: handler,
		},
		{
			name:    "invalid mailbox capacity",
			pool:    pool,
			state:   testWalletState(),
			size:    0,
			handler: handler,
		},
		{
			name:    "nil handler",
			pool:    pool,
			state:   testWalletState(),
			size:    1,
			handler: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := newActor(tt.pool, tt.state, tt.size, tt.handler); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestActorProcessesMessagesInOrder 验证同一个 Actor 按 FIFO 顺序处理消息
func TestActorProcessesMessagesInOrder(t *testing.T) {
	pool := newTestWorkerPool(t, 3)

	got := make([]int, 0, 3)
	a := newTestActor(t, pool, 3, func(state *State, message any) {
		got = append(got, message.(int))
		state.Version++
	})

	for _, message := range []int{1, 2, 3} {
		if err := a.enqueue(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}

	a.close()
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

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
		started <- struct{}{}
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

	for i := 0; i < 2; i++ {
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
	<-started

	for _, message := range []int{2, 3} {
		if err := a.enqueue(context.Background(), message); err != nil {
			t.Fatal(err)
		}
	}

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
	<-started

	if err := a.enqueue(context.Background(), 2); err != nil {
		t.Fatal(err)
	}

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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := a.wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	a.close()
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
