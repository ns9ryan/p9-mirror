// ============================================================================
// 钱包 Worker Pool 测试
//
// 测试 Worker Pool 的创建、任务调度和关闭行为
// 重点验证关闭状态、任务执行期间的关闭保护以及重复关闭
// 同时验证单个 Actor 故障不会影响其他 Actor 的正常调度
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

// TestWorkerPoolContinuesAfterActorFailure 验证一个 Actor 故障不会影响其他 Actor
//
// 测试流程:
//  1. 使用一个 Worker 创建两个 Actor
//  2. 第一个 Actor 开始处理消息后暂停
//  3. 向第二个 Actor 提交消息, 使其进入 Ready Queue
//  4. 恢复第一个 Actor, 模拟 Kafka 写入结果不明确并触发故障
//  5. 验证第一个 Actor 退出后, Worker 仍能继续处理第二个 Actor
//  6. 关闭两个 Actor 并正常关闭 Worker Pool
func TestWorkerPoolContinuesAfterActorFailure(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	firstStarted := make(chan struct{})
	release := make(chan struct{})
	secondProcessed := make(chan struct{})

	failureErr := errors.New("kafka write result uncertain")

	var first *actor

	// 第一个 Actor 触发故障, 第二个 Actor 正常处理消息
	first, err := newActor(pool, testWalletState(), 1, func(_ *State, _ any) {
		close(firstStarted)

		// 暂停第一个 Actor, 确保第二个 Actor 已进入 Ready Queue
		<-release

		// 模拟 Kafka 写入结果不明确
		first.fail(failureErr)
	})
	if err != nil {
		t.Fatal(err)
	}

	second, err := newActor(pool, testWalletState(), 1, func(_ *State, _ any) {
		// 第二个 Actor 能执行到这里, 说明 Worker Pool 仍在正常调度
		close(secondProcessed)
	})
	if err != nil {
		t.Fatal(err)
	}

	// 测试提前退出时也必须释放第一个 Actor
	var releaseOnce sync.Once
	releaseAll := func() {
		releaseOnce.Do(func() { close(release) })
	}

	// Actor 清理先于 Worker Pool 清理执行
	t.Cleanup(func() {
		first.close()
		if err := first.wait(context.Background()); err != nil &&
			!errors.Is(err, ErrWalletFailed) {
			t.Errorf("wait first actor: %v", err)
		}

		second.close()
		if err := second.wait(context.Background()); err != nil {
			t.Errorf("wait second actor: %v", err)
		}
	})

	// 必须后注册, 利用 Cleanup 后进先出的特性
	// 确保 Handler 先解除阻塞, 再执行 Actor 清理
	t.Cleanup(releaseAll)

	// 先让第一个 Actor 占用唯一的 Worker
	if err := first.enqueue(context.Background(), "first"); err != nil {
		t.Fatal(err)
	}

	// 确保第一个 Actor 已经进入 Handler
	select {
	case <-firstStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("first actor handler did not start")
	}

	// 第一个 Actor 正在执行, 第二个 Actor 只能等待调度
	if err := second.enqueue(context.Background(), "second"); err != nil {
		t.Fatal(err)
	}

	// 恢复第一个 Actor, 让它触发故障并退出
	releaseAll()

	// 第一个 Actor 应以故障状态退出, 并保留原始故障原因
	if err := first.wait(context.Background()); !errors.Is(err, ErrWalletFailed) {
		t.Fatalf("expected ErrWalletFailed, got %v", err)
	}
	if !errors.Is(first.runErr, failureErr) {
		t.Fatalf("expected original failure, got %v", first.runErr)
	}

	// 第一个 Actor 故障退出后, Worker 必须继续执行第二个 Actor
	select {
	case <-secondProcessed:
	case <-time.After(5 * time.Second):
		t.Fatal("worker pool stopped after first actor failure")
	}

	// 第二个 Actor 正常处理完消息后, 主动关闭并等待退出
	second.close()

	if err := second.wait(context.Background()); err != nil {
		t.Fatal(err)
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
