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
//   closed 表示 Actor 是否停止接收新消息, 已接受的消息正常关闭时仍会继续处理
//   failed 表示 Actor 是否发生不可自动恢复的故障
//   stopped 表示 Actor 是否已经完全退出, stopped=true 时 done 已关闭
//
// 生命周期状态:
//
//   空闲:
//     scheduled=false, running=0, closed=false, failed=false, stopped=false
//
//   等待调度:
//     scheduled=true, running=0, closed=false, failed=false, stopped=false
//
//   正在处理:
//     scheduled=true, running>=1, closed=false, failed=false, stopped=false
//
//   正常关闭排空:
//     scheduled=true, running>=1 或 running=0, closed=true, failed=false
//
//   故障终止:
//     closed=true, failed=true, 不再接收或处理新消息
//
//   完全退出:
//     scheduled=false, running=0, closed=true, stopped=true
//
// 调度与并发约束:
//   scheduled 保证同一个 Actor 在 Ready Queue 中最多只有一个有效调度项
//   running 在重新调度交接期间可能大于 1, 但不代表多个 Worker 同时执行 Handler
//   前一个 Worker 必须先结束 Handler, 才能将 Actor 重新放入 Ready Queue
//   State 只由当前 Worker 执行 Handler 时访问, 不需要额外加锁
//   mu 协调消息入队、故障终止、关闭和调度状态转换
//
// 关闭流程:
//   close 停止接收新消息并关闭 Mailbox, 但保留已经入队的消息
//   Worker 继续处理已接受的消息, 队列排空后退出
//
// 故障流程:
//   fail 将 Actor 标记为故障并停止接收新消息
//   已经入队但尚未执行的消息会被拒绝, 不再继续处理
//   当前正在执行的 Handler 结束后, Actor 完成退出
//   故障 Actor 保留在 Shard 中, 防止被重新创建成正常 Actor
//
//   wait 等待 Actor 完全退出, Context 只控制当前等待过程
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/google/uuid"
)

var (
	// ErrActorClosed 表示钱包 Actor 已停止接收新消息
	ErrActorClosed = errors.New("wallet actor is closed")

	// ErrWalletFailed 表示钱包发生故障, 必须人工处理后才能恢复
	ErrWalletFailed = errors.New("wallet actor is in failed state")

	errNilActorContext = errors.New("wallet actor context is nil")
)

// actorHandler 处理一条钱包消息, 并允许修改当前钱包状态, 同一个 Actor 的 Handler 不会并发执行
type actorHandler func(state *State, message any)

// actorFailure 保存首次导致钱包故障的资金事件
type actorFailure struct {
	event walletEvent // 原始资金事件, 包含 operation_id
	cause error       // 首次故障原因
}

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

	scheduled bool // 是否处于调度流程, 包括等待 Worker、处理消息和处理后重新调度
	running   int  // 尚未完成本轮收尾的 Worker 数量, 由 workerStart 和 workerDone 修改
	closed    bool // 是否已停止接收新消息
	failed    bool // 是否发生不可自动恢复的故障
	stopped   bool // 是否已完全退出, 防止重复关闭 done

	// 退出状态
	done    chan struct{} // Actor 完全退出信号, 供 a.wait() 等待
	runErr  error         // Actor 运行期间发生的错误, 由 a.wait() 返回
	failure *actorFailure // 首次故障事件及原因, 由 mu 保护
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

	if a.failed {
		return ErrWalletFailed
	}

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

// ================================ worker执行相关 ================================

// runOne 从 Mailbox 取出并处理一条消息, 返回 true 表示还有消息需要重新调度
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

	// 当前操作发生故障时, 不再消费后续消息
	if a.failed {
		a.scheduled = false
		return false
	}

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

	// 故障或正常关闭后, 等待最后一个 Worker 完成收尾
	if a.closed && !a.scheduled && a.running == 0 {
		a.stopLocked()
	}
}

// ================================ 故障相关 ================================

// fail 将 Actor 标记为故障并拒绝后续消息
func (a *actor) fail(err error) {
	a.failWith(actorFailure{cause: err})
}

// failOperation 保留导致故障的原始资金事件及 operation_id
func (a *actor) failOperation(event walletEvent, err error) {
	event.Ledgers = slices.Clone(event.Ledgers)

	a.failWith(actorFailure{
		event: event,
		cause: err,
	})
}

// failWith 执行不可恢复的故障终止
func (a *actor) failWith(failure actorFailure) {
	if a == nil {
		return
	}

	if failure.cause == nil {
		failure.cause = ErrWalletFailed
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.failed || a.stopped {
		return
	}

	// 先标记故障, 防止后续请求继续进入 Mailbox
	a.failed = true
	a.closed = true

	// 保留首次故障信息, 供后续故障记录和人工核对使用
	a.failure = &failure

	failureErr := failure.cause
	if operationID := failure.event.OperationID; operationID != uuid.Nil {
		failureErr = fmt.Errorf("operation_id=%s: %w", operationID, failureErr)
	}

	// 保留最初的故障原因, 供 wait 返回
	a.runErr = errors.Join(ErrWalletFailed, failureErr)

	// 关闭 Mailbox, 阻止后续消息进入
	a.mailbox.close()

	// 已经入队但尚未执行的消息不能继续处理, 逐条通知调用方, 避免 RPC 一直等待
	for message := range a.mailbox.queue {
		if rejected, ok := message.(interface{ reject(error) }); ok {
			rejected.reject(a.runErr)
		}
	}

	// 当前没有 Worker 任务时, 可以直接结束 Actor
	if !a.scheduled && a.running == 0 {
		a.stopLocked()
	}
}

// failureInfo 获取首次故障信息
func (a *actor) failureInfo() (actorFailure, bool) {
	if a == nil {
		return actorFailure{}, false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.failure == nil {
		return actorFailure{}, false
	}

	failure := *a.failure
	failure.event.Ledgers = slices.Clone(failure.event.Ledgers)

	return failure, true
}

// ================================ 退出相关 ================================

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

	// 当前没有待调度消息和 Worker 任务, 可以直接结束 Actor
	if !a.scheduled && a.running == 0 {
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
	if a.stopped || !a.closed || a.scheduled || a.running != 0 {
		return
	}

	a.stopped = true
	close(a.done)
}
