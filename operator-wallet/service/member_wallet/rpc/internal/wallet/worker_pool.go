// ============================================================================
// 钱包 Worker Pool
//
// Worker Pool 管理固定数量的 Worker, 通过 Ready Queue 调度钱包 Actor
// 每个 Worker 每次只处理一个 Actor 的一条消息, 处理后再决定是否重新调度
//
// 调度流程:
//   Actor.enqueue -> schedule -> Ready Queue -> Worker.next
//   -> Actor.runOne -> 重新调度 / 完成本轮处理
//
// 调度约定:
//   Ready Queue 只保存 Actor 指针, 不保存业务消息
//   Actor.scheduled 保证同一个 Actor 最多只有一个有效调度项
//   Actor 每次处理一条消息后重新进入队尾, 让其他 Actor 有机会先执行
//
// Worker 状态:
//   active 表示已经从 Ready Queue 取出但尚未完成本轮处理的任务数
//   next 取出任务时递增, Worker 完成重新调度后递减
//   active 与 Actor.running 不同, 前者统计整个 Worker Pool, 后者统计单个 Actor
//
// 关闭流程:
//   先停止接收新请求, 关闭并等待所有 Actor
//   再关闭 Worker Pool, 如果仍有待调度或正在执行的任务则返回 ErrWorkerPoolBusy
//   Worker Pool 关闭后唤醒所有空闲 Worker, 并等待它们退出
// ============================================================================

package wallet

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrWorkerPoolClosed 表示钱包 Worker Pool 已关闭
	ErrWorkerPoolClosed = errors.New("wallet worker pool is closed")

	// ErrWorkerPoolBusy 表示仍有 Actor 待调度或正在执行
	ErrWorkerPoolBusy = errors.New("wallet worker pool is busy")

	errNilWorkerActor = errors.New("wallet worker actor is nil")
)

// workerPool 使用固定数量的 Worker 调度钱包 Actor
type workerPool struct {
	mu   sync.Mutex // 保护调度队列和 Worker Pool 状态
	cond *sync.Cond // 用于通知等待中的 Worker 重新检查队列和关闭状态

	// Ready Queue
	ready     []*actor // 按 FIFO 顺序保存待处理的 Actor
	readyHead int      // 下一个待取出的 Actor 位置

	// Worker 状态
	active  int            // 已取出但尚未完成本轮处理的任务数
	closed  bool           // 是否已停止接收调度任务
	workers sync.WaitGroup // 等待所有 Worker 退出
}

// newWorkerPool 创建指定 Worker 数量的共享调度池
func newWorkerPool(workerCount int) (*workerPool, error) {
	if workerCount <= 0 {
		return nil, fmt.Errorf("create wallet worker pool: invalid worker count: %d", workerCount)
	}

	p := &workerPool{}
	p.cond = sync.NewCond(&p.mu)

	// 先登记全部 Worker, 再启动 Goroutine
	p.workers.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go p.worker()
	}

	return p, nil
}

// schedule 将一个 Actor 放入 Ready Queue
// 由 Actor 首次进入调度流程或处理完消息后调用
func (p *workerPool) schedule(a *actor) error {
	if p == nil {
		return ErrWorkerPoolClosed
	}

	if a == nil {
		return errNilWorkerActor
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrWorkerPoolClosed
	}

	// 放入队尾, 保持不同 Actor 之间的 FIFO 调度顺序
	p.ready = append(p.ready, a)

	// 只需唤醒一个 Worker, 避免多个 Worker 同时争抢一个任务
	p.cond.Signal()

	return nil
}

// close 关闭 Worker Pool 并等待所有 Worker 退出
// 调用前必须先关闭并等待所有 Actor
func (p *workerPool) close() error {
	if p == nil {
		return nil
	}

	p.mu.Lock()

	if p.closed {
		// 可能已有其他调用开始关闭流程, 仍需等待所有 Worker 退出
		p.mu.Unlock()
		p.workers.Wait()
		return nil
	}

	// 仍有任务时拒绝关闭, 避免 Worker 退出后留下未处理的 Actor
	if p.active > 0 || p.readyHead < len(p.ready) {
		p.mu.Unlock()
		return ErrWorkerPoolBusy
	}

	p.closed = true

	// 唤醒所有空闲 Worker, 让它们检查关闭状态并退出
	p.cond.Broadcast()
	p.mu.Unlock()

	// 不持有调度锁等待, 避免影响 Worker 退出
	p.workers.Wait()

	return nil
}

// worker 持续获取 Actor 并处理一条消息
func (p *workerPool) worker() {
	defer p.workers.Done()

	for {
		a, ok := p.next()
		if !ok {
			return
		}

		// 记录当前 Actor 的 Worker 执行任务, 防止 Actor 过早退出
		a.workerStart()

		// 处理一条消息, 有剩余消息时需要重新调度
		if a.runOne() {
			// 当前消息已经处理完毕, 将 Actor 放回队尾
			if err := p.schedule(a); err != nil {
				// 当前任务尚未释放 active, Worker Pool 此时不允许关闭
				// 因此正常关闭流程下, 重新调度不应失败
				// 出现异常时停止 Actor, 避免它一直处于 scheduled 状态
				a.mu.Lock()
				a.closed = true
				a.mailbox.close()
				a.scheduled = false

				if a.runErr == nil {
					a.runErr = fmt.Errorf("reschedule wallet actor: %w", err)
				}
				a.mu.Unlock()
			}
		}

		// 当前轮处理及重新调度已经结束, 释放 Worker Pool 任务计数
		p.finishTask()

		// 释放当前 Actor 的 Worker 计数, 满足退出条件时通知等待方
		a.workerDone()
	}
}

// next 获取下一个待处理的 Actor
// 没有任务时等待唤醒, Worker Pool 关闭且队列为空时退出
func (p *workerPool) next() (*actor, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Wait 会暂时释放互斥锁, 被唤醒后重新获取锁
	// 使用 for 重新检查条件, 避免任务已被其他 Worker 取走
	for p.readyHead >= len(p.ready) && !p.closed {
		p.cond.Wait()
	}

	// Worker Pool 已关闭且没有剩余任务
	if p.readyHead >= len(p.ready) {
		return nil, false
	}

	a := p.ready[p.readyHead]

	// 清除已消费的引用, 避免底层数组继续持有 Actor
	p.ready[p.readyHead] = nil
	p.readyHead++

	// 从取出任务开始计数, 直到本轮处理和重新调度结束
	p.active++

	p.compactReady()

	return a, true
}

// finishTask 标记一个 Worker 调度任务已完成
func (p *workerPool) finishTask() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.active--
}

// compactReady 回收已经消费的队列空间
// 调用时必须持有 Worker Pool 的互斥锁
func (p *workerPool) compactReady() {
	if p.readyHead == len(p.ready) {
		// 队列已全部消费, 直接重置避免保留历史数据
		p.ready = nil
		p.readyHead = 0
		return
	}

	// 已消费数量不足 64 个或尚未达到队列的一半时, 暂不压缩, 减少频繁复制
	if p.readyHead < 64 || p.readyHead*2 < len(p.ready) {
		return
	}

	// 将尚未消费的 Actor 移到队首, 保持原有 FIFO 顺序
	remaining := copy(p.ready, p.ready[p.readyHead:])

	// 清除尾部不再使用的 Actor 引用
	clear(p.ready[remaining:])

	// 缩短切片长度, 保留原有底层数组
	p.ready = p.ready[:remaining]

	// 从队首重新开始消费
	p.readyHead = 0
}
