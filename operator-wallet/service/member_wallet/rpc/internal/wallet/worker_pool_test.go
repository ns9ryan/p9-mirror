// ============================================================================
// 钱包 Worker Pool 测试
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestNewWorkerPoolInvalidWorkerCount 验证 Worker 数量必须大于零
func TestNewWorkerPoolInvalidWorkerCount(t *testing.T) {
	for _, workerCount := range []int{0, -1} {
		if _, err := newWorkerPool(workerCount); err == nil {
			t.Errorf("expected error for worker count %d", workerCount)
		}
	}
}

// TestWorkerPoolRejectsScheduleAfterClose 验证关闭后拒绝新的 Actor 调度
func TestWorkerPoolRejectsScheduleAfterClose(t *testing.T) {
	pool, err := newWorkerPool(1)
	if err != nil {
		t.Fatal(err)
	}

	if err := pool.close(); err != nil {
		t.Fatal(err)
	}

	if err := pool.close(); err != nil {
		t.Fatal(err)
	}

	if err := pool.schedule(&actor{}); !errors.Is(err, ErrWorkerPoolClosed) {
		t.Fatalf("expected ErrWorkerPoolClosed, got %v", err)
	}
}

// TestWorkerPoolCloseRejectsActiveActor 验证 Actor 尚未退出时不能关闭 Worker Pool
func TestWorkerPoolCloseRejectsActiveActor(t *testing.T) {
	pool, err := newWorkerPool(1)
	if err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once

	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}

	var a *actor

	defer func() {
		releaseAll()

		if a != nil {
			a.close()
			if err := a.wait(context.Background()); err != nil {
				t.Errorf("wait actor: %v", err)
			}
		}

		if err := pool.close(); err != nil {
			t.Errorf("close worker pool: %v", err)
		}
	}()

	a, err = newActor(pool, testWalletState(), 1, func(_ *State, _ any) {
		close(started)
		<-release
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := a.enqueue(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("actor did not start")
	}

	if err := pool.close(); !errors.Is(err, ErrWorkerPoolBusy) {
		t.Fatalf("expected ErrWorkerPoolBusy, got %v", err)
	}

	a.close()
	releaseAll()

	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := pool.close(); err != nil {
		t.Fatal(err)
	}
}
