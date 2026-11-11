// ============================================================================
// 钱包 Actor 执行与生命周期
//
// Actor 独占管理一个钱包状态, 通过 Mailbox 接收消息, 由共享 Worker Pool 调度
// 每个 Actor 同一时刻最多只有一个 Worker 执行 Handler, 每次只处理一条消息
//
// 消息处理流程:
//   enqueue -> Mailbox -> Worker Pool -> Ready Queue -> Worker
//   -> runOne -> Handler -> 检查 Mailbox -> 重新调度 / 等待新消息
//
// 字段关系 (以下生命周期字段均由 mu 保护):
//   scheduled 表示 Actor 是否处于调度流程, 包括等待 Worker、处理消息和重新调度
//   running 表示已经取得 Actor 但尚未完成本轮收尾的 Worker 数量
//   closed 表示 Actor 是否停止接收新消息, 已接受的消息仍会继续处理
//   stopped 表示 Actor 是否已经完全退出, stopped=true 时 done 已关闭
//
// 生命周期状态:
//
//   空闲: 没有待处理消息, 也没有 Worker 任务
//     scheduled=false, running=0, closed=false, stopped=false
//
//   等待调度: 已有消息进入 Mailbox, Actor 已进入 Ready Queue, 等待 Worker 处理
//     scheduled=true, running=0, closed=false, stopped=false
//
//   正在处理: Worker 已取得 Actor, 正在处理消息
//     scheduled=true, running>=1, closed=false, stopped=false
//
//   处理收尾: 消息处理结束, Mailbox 已排空, 但 Worker 尚未完成本轮收尾
//     scheduled=false, running>=1, closed=false, stopped=false
//
//   关闭排空: Actor 已停止接收新消息, 但已接受的消息仍在等待调度或处理
//     scheduled=true, running=0 或 running>=1, closed=true, stopped=false
//
//   关闭收尾: 已没有待处理消息, 但仍有 Worker 尚未完成本轮收尾
//     scheduled=false, running>=1, closed=true, stopped=false
//
//   完全退出: 没有待处理消息, 所有 Worker 均已完成收尾, Actor 已完全退出
//     scheduled=false, running=0, closed=true, stopped=true
//
// 调度与并发约束:
//   scheduled 保证同一个 Actor 在 Ready Queue 中最多只有一个有效调度项
//   running 在重新调度交接期间可能大于 1, 但不代表多个 Worker 同时执行 Handler
//   前一个 Worker 必须先结束 Handler, 才能将 Actor 重新放入 Ready Queue
//   State 只由当前 Worker 执行 Handler 时访问, 不需要额外加锁
//   mu 协调消息入队、关闭和调度状态转换, 避免消息遗漏或重复调度
//
// 关闭流程:
//   close 停止接收新消息并关闭 Mailbox, 但保留已经入队的消息
//   Worker 继续处理已接受的消息, 队列排空后解除调度状态
//   当 closed=true、scheduled=false 且 running=0 时, 关闭 done 通知等待方
//   wait 等待 Actor 完全退出, Context 只控制当前等待过程
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

// actorHandler 处理一条钱包消息, 并允许修改当前钱包状态
// 同一个 Actor 的 Handler 不会并发执行
type actorHandler func(state *State, message any)

// actor 独占管理一个钱包状态, 消息由共享 Worker Pool 串行处理
type actor struct {
	// 钱包运行时状态
	state State // 当前钱包状态, 仅由 Actor Handler 访问

	// 消息处理依赖
	mailbox *mailbox     // 当前钱包的有界消息队列
	pool    *workerPool  // 负责调度当前 Actor 的共享 Worker Pool
	handler actorHandler // 当前钱包的消息处理逻辑

	// 生命周期状态
	mu sync.Mutex // 保护下面的生命周期字段

	scheduled bool // 是否处于调度流程, 包括在 Ready Queue 等待、Worker 处理消息和处理后重新调度
	running   int  // 尚未完成本轮收尾的 Worker 数量, 由 workerStart 和 workerDone 修改
	closed    bool // 是否已停止接收新消息
	stopped   bool // 是否已完全退出, 防止重复关闭 done

	done   chan struct{} // Actor 完全退出信号, 供 a.wait() 等待
	runErr error         // 记录 Worker Pool 调度失败或 Mailbox 非预期读取错误, 由 a.wait() 返回
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

// enqueue 接收一条消息, 并确保当前 Actor 进入调度流程
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

	// 已经等待调度或正在处理时, 当前消息由已有调度流程负责
	// 再次提交会导致同一个 Actor 在 Ready Queue 中出现多个调度项
	if a.scheduled {
		return nil
	}

	// 先标记调度状态, 再提交 Worker Pool, 避免并发入队时重复调度
	a.scheduled = true

	if err := a.pool.schedule(a); err != nil {
		// scheduled 原本为 false, 说明没有待处理消息或有效调度项
		// 当前消息是刚入队的唯一消息, 调度失败后需要将其移除
		<-a.mailbox.queue

		a.scheduled = false
		a.closed = true
		a.mailbox.close()

		scheduleErr := fmt.Errorf("schedule wallet actor: %w", err)
		if a.runErr == nil {
			a.runErr = scheduleErr
		}

		// 没有尚未完成收尾的 Worker, 可以直接通知 Actor 退出
		if a.running == 0 {
			a.stopLocked()
		}

		// running > 0 时, 由最后一个 workerDone() 负责通知 Actor 退出
		return scheduleErr
	}

	return nil
}

// runOne 从 Mailbox 取出并处理一条消息
// 返回 true 表示还有消息需要重新调度
func (a *actor) runOne() bool {
	// 消息一旦入队就由 Actor 接管, 不再受提交请求的 Context 取消影响
	message, err := a.mailbox.dequeue(context.Background())
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		// Actor 主动关闭 Mailbox 并排空消息后, dequeue 会返回 ErrMailboxClosed
		normalClose := a.closed && errors.Is(err, ErrMailboxClosed)
		if a.runErr == nil && !normalClose {
			a.runErr = fmt.Errorf("dequeue wallet actor message: %w", err)
		}

		// 读取异常且 Actor 尚未关闭时, 停止接收新消息并关闭 Mailbox
		if !a.closed {
			a.closed = true
			a.mailbox.close()
		}

		// 当前消息队列已经无法继续消费, 结束调度流程
		a.scheduled = false
		return false
	}

	// 当前 Worker 独占执行 Handler, 可以直接修改钱包状态
	a.handler(&a.state, message)

	a.mu.Lock()
	defer a.mu.Unlock()

	// enqueue 和 close 同样需要持有 mu, 此处检查队列不会遗漏并发入队
	if len(a.mailbox.queue) > 0 {
		// 仍有消息需要处理, 保持 scheduled=true, 由 Worker 放回 Ready Queue
		return true
	}

	// 队列为空, 解除调度状态, 后续新消息可以重新触发调度
	a.scheduled = false

	return false
}

// workerStart 记录 Worker 开始处理当前 Actor
func (a *actor) workerStart() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 从 Worker 取得 Actor 开始计数, 直到完成本轮收尾
	a.running++
}

// workerDone 记录 Worker 已完成当前 Actor 的本轮处理
func (a *actor) workerDone() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.running--

	// 必须没有待处理消息且所有 Worker 都完成收尾, 才能通知 Actor 退出
	if a.closed && !a.scheduled && a.running == 0 {
		a.stopLocked()
	}
}

// close 停止接收新消息, 并保留已经入队的消息供 Actor 继续处理
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

	// 关闭 Mailbox 后拒绝新消息, 但缓冲区中的消息仍可继续读取
	a.mailbox.close()

	if !a.scheduled && a.running == 0 {
		// 当前没有待调度消息和 Worker 任务, 可以直接结束 Actor
		a.stopLocked()
	}

	// 仍有调度任务时不直接退出, 由最后一个 Worker 完成退出通知
}

// wait 等待 Actor 完全退出, Context 只控制当前等待过程
func (a *actor) wait(ctx context.Context) error {
	if a == nil {
		return ErrActorClosed
	}

	if ctx == nil {
		return errNilActorContext
	}

	// 优先响应调用前已经取消的 Context
	if err := ctx.Err(); err != nil {
		return err
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

// stopLocked 关闭 Actor 退出信号, 调用时必须持有 mu
func (a *actor) stopLocked() {
	if a.stopped {
		return
	}

	a.stopped = true
	close(a.done)
}
