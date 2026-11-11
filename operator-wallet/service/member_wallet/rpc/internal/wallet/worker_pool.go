// ============================================================================
// 钱包 Worker Pool
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
	mu      sync.Mutex     // 保护调度队列和运行状态
	cond    *sync.Cond     // 有 Actor 就绪时唤醒 Worker
	ready   []*actor       // 待调度的 Actor, 每个 Actor 同时最多出现一次
	head    int            // 下一个待调度 Actor 的位置
	active  int            // 当前正在执行的 Actor 数量
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
	p.workers.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go p.worker()
	}

	return p, nil
}

// schedule 将一个待处理 Actor 放入 Ready Queue
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

	p.ready = append(p.ready, a)
	p.cond.Signal()

	return nil
}

// close 关闭 Worker Pool, 仅允许在所有 Actor 退出后调用
func (p *workerPool) close() error {
	if p == nil {
		return nil
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		p.workers.Wait()
		return nil
	}

	// 防止 Actor 尚未处理完消息时关闭 Worker
	if p.active > 0 || p.head < len(p.ready) {
		p.mu.Unlock()
		return ErrWorkerPoolBusy
	}

	p.closed = true
	p.cond.Broadcast()
	p.mu.Unlock()

	p.workers.Wait()
	return nil
}

// worker 持续获取并执行 Ready Queue 中的 Actor
func (p *workerPool) worker() {
	defer p.workers.Done()

	for {
		a, ok := p.next()
		if !ok {
			return
		}

		a.workerStart()
		a.runOne()
		a.prepareNext()
		p.finishTask()
		a.workerDone()
	}
}

// next 获取下一个就绪 Actor, 没有任务时等待唤醒
func (p *workerPool) next() (*actor, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.head >= len(p.ready) && !p.closed {
		p.cond.Wait()
	}

	if p.head >= len(p.ready) {
		return nil, false
	}

	a := p.ready[p.head]
	p.ready[p.head] = nil
	p.head++
	p.active++
	p.compactReady()

	return a, true
}

// finishTask 标记一个 Actor 调度任务已退出
func (p *workerPool) finishTask() {
	p.mu.Lock()
	p.active--
	p.mu.Unlock()
}

// compactReady 回收已消费的队列空间, 避免长期保留无用引用
func (p *workerPool) compactReady() {
	if p.head == len(p.ready) {
		p.ready = nil
		p.head = 0
		return
	}

	if p.head < 64 || p.head*2 < len(p.ready) {
		return
	}

	remaining := copy(p.ready, p.ready[p.head:])
	clear(p.ready[remaining:])
	p.ready = p.ready[:remaining]
	p.head = 0
}
