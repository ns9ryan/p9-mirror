// ============================================================================
// 钱包 Worker Pool 测试
//
// 测试 Worker Pool 的创建、任务调度和关闭行为
// 重点验证关闭状态、任务执行期间的关闭保护以及重复关闭
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// ---- Worker Pool 创建与调度 ----

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

	// 空闲 Worker Pool 可以正常关闭
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}

	// 重复关闭应等待 Worker 退出, 不应产生错误
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}

	// Worker Pool 已关闭, 不再允许提交新的 Actor
	if err := pool.schedule(&actor{}); !errors.Is(err, ErrWorkerPoolClosed) {
		t.Fatalf("expected ErrWorkerPoolClosed, got %v", err)
	}
}

// ---- Worker Pool 关闭 ----

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
		// 测试正常结束和提前退出都可能释放 Handler, 只允许关闭一次
		releaseOnce.Do(func() { close(release) })
	}

	var a *actor

	defer func() {
		// 即使测试提前失败, 也先释放 Handler, 避免清理过程一直等待
		releaseAll()

		if a != nil {
			a.close()
			if err := a.wait(context.Background()); err != nil {
				t.Errorf("wait actor: %v", err)
			}
		}

		// Actor 退出后再关闭 Worker Pool
		if err := pool.close(); err != nil {
			t.Errorf("close worker pool: %v", err)
		}
	}()

	a, err = newActor(pool, testWalletState(), 1, func(_ *State, _ any) {
		// 通知测试 Handler 已经开始执行
		close(started)

		// 暂停 Handler, 保持 Worker Pool 的 active 大于零
		<-release
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := a.enqueue(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}

	// 确保 Actor 已经被 Worker 取出并进入 Handler
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("actor did not start")
	}

	// Handler 尚未结束, Worker Pool 仍有活动任务, 此时必须拒绝关闭
	if err := pool.close(); !errors.Is(err, ErrWorkerPoolBusy) {
		t.Fatalf("expected ErrWorkerPoolBusy, got %v", err)
	}

	// 停止接收新消息, 然后释放正在执行的 Handler
	a.close()
	releaseAll()

	// 等待 Actor 完成本轮处理并退出
	if err := a.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Actor 已退出且没有剩余调度任务, 此时允许关闭 Worker Pool
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}
}
