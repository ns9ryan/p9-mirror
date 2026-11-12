// ============================================================================
// 钱包 Shard 测试
//
// 测试 Shard 的 Actor 创建、复用、身份校验和生命周期
// 重点验证并发创建的唯一性，以及关闭期间不会遗漏已创建的 Actor
//
// 并发创建:
//   多个 Goroutine 同时创建同一个钱包, 最终只能得到一个 Actor
//   创建与关闭同时发生时, 请求要么成功创建, 要么返回 ErrShardClosed
//
// 关闭流程:
//   Shard.close 停止创建 Actor, 并关闭所有已登记的 Actor
//   Actor 已接受的消息仍由 Worker Pool 继续处理
//   Shard.wait 等待全部 Actor 退出, Context 只控制当前等待过程
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ---- 测试辅助方法 ----

// newTestShard 创建测试 Shard, 并在测试结束时关闭和等待所有 Actor
func newTestShard(t *testing.T, pool *workerPool, handler actorHandler) *shard {
	t.Helper()

	s, err := newShard(0, pool, 2, handler)
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup 按后进先出执行
	// Shard 的 Cleanup 后注册, 因此会先于 Worker Pool 的 Cleanup 执行
	// 确保全部 Actor 退出后, 才关闭共享 Worker Pool
	t.Cleanup(func() {
		s.close()
		if err := s.wait(context.Background()); err != nil {
			t.Errorf("wait shard: %v", err)
		}
	})

	return s
}

// ---- 创建与复用 ----

// TestNewShardRejectsInvalidConfig 验证 Shard 拒绝无效配置
func TestNewShardRejectsInvalidConfig(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	handler := actorHandler(func(_ *State, _ any) {})

	tests := []struct {
		name            string       // 测试场景
		index           int          // Shard 下标
		pool            *workerPool  // 共享 Worker Pool
		mailboxCapacity int          // Actor Mailbox 容量
		handler         actorHandler // Actor 消息处理函数
	}{
		{"negative index", -1, pool, 1, handler},
		{"nil worker pool", 0, nil, 1, handler},
		{"invalid mailbox capacity", 0, pool, 0, handler},
		{"nil handler", 0, pool, 1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := newShard(tt.index, tt.pool, tt.mailboxCapacity, tt.handler); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestShardGetOrCreateReusesActor 验证相同钱包只使用一个 Actor, 且不会覆盖运行时状态
func TestShardGetOrCreateReusesActor(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	s := newTestShard(t, pool, func(_ *State, _ any) {})
	state := testWalletState()

	// 首次查询时, 钱包尚未创建 Actor
	if a, err := s.get(state.WalletGlobalID); err != nil || a != nil {
		t.Fatalf("expected missing actor, got actor=%v, err=%v", a, err)
	}

	// 首次调用 getOrCreate, 使用传入的 State 创建 Actor
	first, err := s.getOrCreate(state)
	if err != nil {
		t.Fatal(err)
	}

	// 模拟再次加载到不同的余额和版本
	// 已存在的 Actor 必须继续使用自己的运行时 State, 不能被旧快照覆盖
	updated := state
	updated.Balance = 100
	updated.Version = 2

	second, err := s.getOrCreate(updated)
	if err != nil {
		t.Fatal(err)
	}

	// 相同 WalletGlobalID 必须返回同一个 Actor
	if first != second {
		t.Fatal("same wallet created multiple actors")
	}
	if first.state.Balance != 0 || first.state.Version != 0 {
		t.Fatal("existing actor state was overwritten")
	}

	// get 只能查找已有 Actor, 返回的实例也必须一致
	found, err := s.get(state.WalletGlobalID)
	if err != nil || found != first {
		t.Fatalf("expected existing actor, got actor=%v, err=%v", found, err)
	}
}

// TestShardConcurrentGetOrCreate 验证并发请求不会重复创建同一个钱包 Actor
//
// 测试流程:
//  1. 创建 64 个 Goroutine, 同时等待启动信号
//  2. 统一释放启动信号, 让所有 Goroutine 竞争创建同一个钱包
//  3. 收集全部返回结果, 检查 Actor 指针是否一致
//  4. 确认 Shard 的 Actor Map 最终只有一条记录
func TestShardConcurrentGetOrCreate(t *testing.T) {
	pool := newTestWorkerPool(t, 4)
	s := newTestShard(t, pool, func(_ *State, _ any) {})
	state := testWalletState()

	const requestCount = 64

	type result struct {
		actor *actor // 创建或获取到的 Actor
		err   error  // 创建或获取时的错误
	}

	// start 用于统一触发并发请求
	// results 使用缓冲 Channel, 容量等于请求数量, 避免发送结果时阻塞
	start := make(chan struct{})
	results := make(chan result, requestCount)

	// 提前启动所有 Goroutine, 让它们等待同一个关闭信号
	for i := 0; i < requestCount; i++ {
		go func() {
			<-start

			a, err := s.getOrCreate(state)
			results <- result{actor: a, err: err}
		}()
	}

	// 关闭 Channel 会同时唤醒所有等待 start 的 Goroutine
	// 此时它们会并发调用 getOrCreate, 竞争创建同一个钱包 Actor
	close(start)

	// 收集全部结果, 验证所有请求返回的都是同一个 Actor
	var first *actor
	for i := 0; i < requestCount; i++ {
		got := <-results
		if got.err != nil {
			t.Fatal(got.err)
		}
		if first == nil {
			first = got.actor
		}
		if got.actor != first {
			t.Fatal("concurrent requests created multiple actors")
		}
	}

	// 即使有 64 个并发请求, Shard 内也只能登记一个 Actor
	if len(s.actors) != 1 {
		t.Fatalf("expected one actor, got %d", len(s.actors))
	}
}

// TestShardRejectsWalletIdentityMismatch 验证同一个钱包不能跨归属复用 Actor
func TestShardRejectsWalletIdentityMismatch(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	s := newTestShard(t, pool, func(_ *State, _ any) {})
	state := testWalletState()

	// 先创建合法的钱包 Actor, 作为后续身份冲突测试的基准
	if _, err := s.getOrCreate(state); err != nil {
		t.Fatal(err)
	}

	// 保持 WalletGlobalID 不变, 分别修改其他不可变身份字段
	// 这些字段发生冲突时, 不允许复用原有 Actor
	tests := []struct {
		name   string       // 身份冲突场景
		change func(*State) // 修改指定的钱包身份字段
	}{
		{"wallet ID", func(s *State) { s.WalletID++ }},
		{"member", func(s *State) { s.MemberGlobalID = uuid.New() }},
		{"operator", func(s *State) { s.OperatorCode = "op002" }},
		{"currency", func(s *State) { s.CurrencyCode = "CNY" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := state
			tt.change(&changed)

			if _, err := s.getOrCreate(changed); !errors.Is(err, ErrShardWalletIdentityMismatch) {
				t.Fatalf("expected ErrShardWalletIdentityMismatch, got %v", err)
			}
		})
	}
}

// ---- 关闭与等待 ----

// TestShardCloseDrainsActors 验证关闭 Shard 后所有 Actor 会排空已接受的消息
//
// 测试流程:
//  1. 让 Actor1 开始处理消息 1, 并暂停在 Handler 内
//  2. 向 Actor1 追加消息 2, 向 Actor2 提交消息 3
//  3. 关闭 Shard, 验证不再接受新消息
//  4. 恢复 Actor1 执行, 等待全部 Actor 退出
//  5. 验证关闭前接受的 3 条消息全部处理完毕
func TestShardCloseDrainsActors(t *testing.T) {
	pool := newTestWorkerPool(t, 2)

	// started 通知测试消息 1 已经进入 Handler
	// release 控制消息 1 何时继续执行
	// processed 记录已经处理完成的消息
	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan int, 3)

	s := newTestShard(t, pool, func(_ *State, message any) {
		value := message.(int)

		if value == 1 {
			close(started)

			// 暂停 Actor1, 确保消息 2 能留在 Mailbox 中
			<-release
		}

		processed <- value
	})

	// 正常执行和测试提前退出都可能释放 Handler
	// 使用 sync.Once 防止重复 close(release) 引发 panic
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)

	// 创建两个不同钱包的 Actor, 共用同一个 Worker Pool
	state1 := testWalletState()
	state2 := testWalletState()
	state2.WalletID = 2

	a1, err := s.getOrCreate(state1)
	if err != nil {
		t.Fatal(err)
	}

	a2, err := s.getOrCreate(state2)
	if err != nil {
		t.Fatal(err)
	}

	// 提交消息 1, 并确保 Actor1 已进入暂停状态
	if err := a1.enqueue(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first actor handler did not start")
	}

	// Actor1 正在处理消息 1, 消息 2 只能等待在 Mailbox 中
	// Actor2 是独立钱包, 可以由其他 Worker 处理消息 3
	if err := a1.enqueue(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	if err := a2.enqueue(context.Background(), 3); err != nil {
		t.Fatal(err)
	}

	// 关闭 Shard, 停止接收新消息, 但保留此前已经接受的消息
	// 重复 close 不应产生副作用
	s.close()
	s.close()

	// Shard 关闭后不能再获取或创建 Actor
	if _, err := s.getOrCreate(state1); !errors.Is(err, ErrShardClosed) {
		t.Fatalf("expected ErrShardClosed, got %v", err)
	}

	// 已关闭的 Actor 同样不允许继续入队
	if err := a1.enqueue(context.Background(), 4); !errors.Is(err, ErrActorClosed) {
		t.Fatalf("expected ErrActorClosed, got %v", err)
	}

	// 释放 Actor1, 让消息 1 结束并继续处理消息 2
	releaseAll()

	// Shard.wait 必须等待所有 Actor 完成消息处理后才能返回
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.wait(ctx); err != nil {
		t.Fatal(err)
	}

	// 不同 Actor 可以并发执行, 因此不要求跨钱包的完成顺序
	// 只验证关闭前接受的消息 1、2、3 都已经处理
	seen := make(map[int]bool)
	for i := 0; i < 3; i++ {
		seen[<-processed] = true
	}

	for _, message := range []int{1, 2, 3} {
		if !seen[message] {
			t.Fatalf("message %d was not processed", message)
		}
	}
}

// TestShardWaitRequiresClose 验证未关闭的 Shard 不允许等待退出
func TestShardWaitRequiresClose(t *testing.T) {
	pool := newTestWorkerPool(t, 1)
	s := newTestShard(t, pool, func(_ *State, _ any) {})

	if err := s.wait(context.Background()); !errors.Is(err, ErrShardNotClosed) {
		t.Fatalf("expected ErrShardNotClosed, got %v", err)
	}
}

// TestShardWaitContextCanceled 验证取消等待不影响 Shard 的实际关闭流程
//
// 测试流程:
//  1. 暂停正在处理消息的 Actor
//  2. 关闭 Shard, 但 Actor 尚未完成退出
//  3. 使用已取消的 Context 调用 wait, 验证立即返回取消错误
//  4. 确认 Actor 没有被取消的 Context 强制结束
//  5. 释放 Handler, 再次等待 Shard 正常退出
func TestShardWaitContextCanceled(t *testing.T) {
	pool := newTestWorkerPool(t, 1)

	started := make(chan struct{})
	release := make(chan struct{})

	s := newTestShard(t, pool, func(_ *State, _ any) {
		// 通知测试 Handler 已经开始执行, 然后等待释放信号
		close(started)
		<-release
	})

	// 确保测试提前退出时也能释放正在等待的 Handler
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)

	a, err := s.getOrCreate(testWalletState())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.enqueue(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}

	// 等待 Handler 开始, 确保 Actor 正在运行而不是已经退出
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("actor handler did not start")
	}

	// 关闭 Shard 后, Actor 仍需处理完已经接受的消息
	s.close()

	// 预先取消 Context, 验证 wait 会返回 context.Canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := s.wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// Context 取消只结束本次等待, 不应该让 Actor 提前退出
	select {
	case <-a.done:
		t.Fatal("actor exited before pending handler completed")
	default:
	}

	// 恢复 Handler, 再使用正常 Context 等待实际关闭完成
	releaseAll()
	if err := s.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// ---- 并发创建与关闭 ----

// TestShardConcurrentCreateAndClose 验证并发创建和关闭不会遗漏 Actor
//
// 测试流程:
//  1. 准备 64 个不同钱包的创建请求
//  2. 同时释放创建请求并触发 Shard.close
//  3. 验证每个创建请求只可能成功或返回 ErrShardClosed
//  4. 等待 Shard 完全退出, 确认关闭后不再允许访问
//
// 核心验证:
//
//	创建先取得 Shard.mu 时, Actor 会完成登记, 后续由 close 负责关闭
//	close 先取得 Shard.mu 时, 后续创建请求会返回 ErrShardClosed
//	两种执行顺序都不能导致 Actor 遗漏或 Shard 关闭后继续创建
func TestShardConcurrentCreateAndClose(t *testing.T) {
	pool := newTestWorkerPool(t, 4)
	s := newTestShard(t, pool, func(_ *State, _ any) {})

	const requestCount = 64

	// start 统一触发创建请求
	// results 收集每个 Goroutine 的执行结果
	start := make(chan struct{})
	results := make(chan error, requestCount)

	// 每个 Goroutine 使用不同的钱包标识, 模拟不同钱包同时进入 Shard
	for i := 0; i < requestCount; i++ {
		state := testWalletState()
		state.WalletID = int64(i + 1)

		go func() {
			<-start

			_, err := s.getOrCreate(state)
			results <- err
		}()
	}

	// 同时启动创建请求和关闭操作, 制造锁竞争
	// 实际获取 Shard.mu 的顺序由 Goroutine 调度决定
	close(start)
	go s.close()

	// 创建请求要么在关闭前成功, 要么在关闭后被拒绝
	// 其他错误说明并发创建或关闭流程存在异常
	for i := 0; i < requestCount; i++ {
		select {
		case err := <-results:
			if err != nil && !errors.Is(err, ErrShardClosed) {
				t.Fatalf("unexpected create error: %v", err)
			}

		case <-time.After(5 * time.Second):
			t.Fatal("concurrent create did not finish")
		}
	}

	// close 可以重复调用, 并确保全部已创建的 Actor 都进入关闭流程
	s.close()

	// wait 返回后, 所有已登记 Actor 都应该退出
	if err := s.wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Shard 关闭后, 即使只是查询 Actor 也必须被拒绝
	if _, err := s.get(uuid.New()); !errors.Is(err, ErrShardClosed) {
		t.Fatalf("expected ErrShardClosed, got %v", err)
	}
}
