// ============================================================================
// 钱包 Actor 执行与生命周期
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrActorClosed 表示钱包 Actor 已停止接收新消息
	ErrActorClosed = errors.New("wallet actor is closed")

	errNilActorContext = errors.New("wallet actor context is nil")
)

// actorHandler 处理一条钱包消息, 并在 Actor 内更新钱包状态
type actorHandler func(state *State, message any)

// actor 独占管理一个钱包状态, 消息由共享 Worker Pool 串行处理
type actor struct {
	state   State        // 当前钱包运行时状态, 仅由 Actor Handler 访问
	mailbox *mailbox     // 当前钱包的有界消息队列
	pool    *workerPool  // 共享 Worker Pool
	handler actorHandler // 钱包消息处理逻辑

	mu        sync.Mutex    // 保护调度和关闭状态
	scheduled bool          // 是否已进入 Ready Queue 或正在执行
	running   int           // 当前执行该 Actor 的 Worker 数量
	closed    bool          // 是否已停止接收新消息
	stopped   bool          // 是否已处理完消息并退出
	done      chan struct{} // Actor 退出信号
	runErr    error         // Actor 退出前发生的调度错误
}

// newActor 创建钱包 Actor, 并绑定独立的钱包状态和消息队列
func newActor(pool *workerPool, state State, mailboxCapacity int, handler actorHandler) (*actor, error) {
	if pool == nil {
		return nil, errors.New("create wallet actor: worker pool is nil")
	}

	if handler == nil {
		return nil, errors.New("create wallet actor: handler is nil")
	}

	if err := state.Validate(); err != nil {
		return nil, fmt.Errorf("create wallet actor: %w", err)
	}

	m, err := newMailbox(mailboxCapacity)
	if err != nil {
		return nil, fmt.Errorf("create wallet actor mailbox: %w", err)
	}

	return &actor{
		state:   state,
		mailbox: m,
		pool:    pool,
		handler: handler,
		done:    make(chan struct{}),
	}, nil
}

// enqueue 接收一条消息并确保 Actor 进入调度队列
func (a *actor) enqueue(ctx context.Context, message any) error {
	if a == nil {
		return ErrActorClosed
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return ErrActorClosed
	}

	if err := a.mailbox.enqueue(ctx, message); err != nil {
		return err
	}

	if a.scheduled {
		return nil
	}

	a.scheduled = true
	if err := a.pool.schedule(a); err != nil {
		a.scheduled = false

		// 未调度时队列原本为空, 调度失败后移除本次刚入队的消息
		<-a.mailbox.queue
		return err
	}

	return nil
}

// runOne 从邮箱取出并处理一条消息
func (a *actor) runOne() {
	// 消息一旦入队就由 Actor 负责处理, 不再受提交请求的 Context 取消影响
	message, err := a.mailbox.dequeue(context.Background())
	if err != nil {
		a.finish(err)
		return
	}

	a.handler(&a.state, message)
}

// prepareNext 有待处理消息时将 Actor 放回 Ready Queue
func (a *actor) prepareNext() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.mailbox.queue) == 0 {
		a.scheduled = false
		return
	}

	if err := a.pool.schedule(a); err != nil {
		a.scheduled = false
		a.closed = true
		a.mailbox.close()
		a.runErr = fmt.Errorf("reschedule wallet actor: %w", err)
	}
}

// finish 处理消息队列意外结束的情况
func (a *actor) finish(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.scheduled = false

	if !a.closed {
		a.closed = true
		a.mailbox.close()
		a.runErr = fmt.Errorf("wallet actor mailbox closed unexpectedly: %w", err)
		return
	}

	if !errors.Is(err, ErrMailboxClosed) {
		a.runErr = fmt.Errorf("dequeue wallet actor message: %w", err)
	}
}

// workerStart 记录 Worker 开始执行当前 Actor
func (a *actor) workerStart() {
	a.mu.Lock()
	a.running++
	a.mu.Unlock()
}

// workerDone 记录 Worker 已退出, 并在 Actor 满足退出条件时发出信号
func (a *actor) workerDone() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.running--
	if a.closed && !a.scheduled && a.running == 0 {
		a.stopLocked()
	}
}

// close 停止接收新消息, 并保留已入队消息供 Actor 继续处理
func (a *actor) close() {
	if a == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}

	a.closed = true
	a.mailbox.close()

	if !a.scheduled && a.running == 0 {
		a.stopLocked()
	}
}

// wait 等待 Actor 退出, Context 只控制等待过程
func (a *actor) wait(ctx context.Context) error {
	if a == nil {
		return ErrActorClosed
	}

	if ctx == nil {
		return errNilActorContext
	}

	select {
	case <-a.done:
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.runErr

	case <-ctx.Done():
		return ctx.Err()
	}
}

// stopLocked 标记 Actor 已退出并通知等待方, 调用时必须持有互斥锁
func (a *actor) stopLocked() {
	if a.stopped {
		return
	}

	a.stopped = true
	close(a.done)
}
