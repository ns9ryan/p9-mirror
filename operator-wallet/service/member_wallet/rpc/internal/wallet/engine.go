// ============================================================================
// 钱包 Engine 管理与生命周期
//
// Engine 负责组织 Router、Shard 和共享 Worker Pool
// 一个 Engine 管理固定数量的 Shard, 每个钱包由唯一的 Shard 管理
//
// 钱包访问流程:
//   WalletGlobalID -> Router -> Shard -> Actor
//
// 并发约束:
//   mu 保护 Engine 关闭状态, 协调钱包访问和整体关闭
//   多个钱包请求可以并发访问 Shard
//   close 持有写锁, 确保关闭过程中不会创建新的 Actor
//
// 关闭流程:
//   close 停止接收新请求, 并关闭所有 Shard
//   Shard 负责关闭并等待各自管理的 Actor
//   wait 等待全部 Shard 退出后, 关闭共享 Worker Pool
//   Context 取消只结束当前等待, 不会中断已接受的消息处理
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var (
	// ErrEngineClosed 表示 Engine 已停止接收钱包请求
	ErrEngineClosed = errors.New("wallet engine is closed")

	// ErrEngineNotClosed 表示 Engine 尚未进入关闭流程
	ErrEngineNotClosed = errors.New("wallet engine is not closed")

	// ErrWalletActorNotFound 表示钱包 Actor 尚未初始化
	ErrWalletActorNotFound = errors.New("wallet actor not found")

	errNilEngine        = errors.New("wallet engine is nil")
	errNilEngineContext = errors.New("wallet engine context is nil")
)

// engine 统一管理钱包路由、Shard 和共享 Worker Pool
type engine struct {
	router *router     // 根据钱包全局标识计算 Shard 下标
	shards []*shard    // 固定数量的 Shard, 下标与 Router 结果一致
	pool   *workerPool // 所有 Shard 共享的 Worker Pool

	mu     sync.RWMutex // 保护 closed, 协调钱包访问和关闭
	closed bool         // 是否已停止接收新的钱包请求
}

// newEngine 创建钱包 Engine 和全部 Shard, 钱包 Actor 按需创建
func newEngine(shardCount, workerCount, mailboxCapacity int, handler actorHandler) (*engine, error) {
	r, err := newRouter(shardCount)
	if err != nil {
		return nil, fmt.Errorf("create wallet engine router: %w", err)
	}

	// 先校验配置, 避免 Worker Pool 启动后才发现参数无效
	if workerCount <= 0 {
		return nil, fmt.Errorf("create wallet engine: invalid worker count: %d", workerCount)
	}
	if mailboxCapacity <= 0 {
		return nil, fmt.Errorf("create wallet engine: invalid mailbox capacity: %d", mailboxCapacity)
	}
	if handler == nil {
		return nil, errors.New("create wallet engine: handler is nil")
	}

	pool, err := newWorkerPool(workerCount)
	if err != nil {
		return nil, fmt.Errorf("create wallet engine worker pool: %w", err)
	}

	// Shard 数量固定, 每个 Shard 共享同一个 Worker Pool
	shards := make([]*shard, shardCount)
	for i := range shards {
		s, err := newShard(i, pool, mailboxCapacity, handler)
		if err != nil {
			// 此时尚未创建 Actor, 可以直接关闭 Worker Pool
			_ = pool.close()
			return nil, fmt.Errorf("create wallet engine shard %d: %w", i, err)
		}

		shards[i] = s
	}

	return &engine{
		router: r,      // 固定钱包路由规则
		shards: shards, // 按下标管理全部 Shard
		pool:   pool,   // 所有钱包共享的 Worker Pool
	}, nil
}

// get 根据钱包全局标识查找 Actor, 不存在时返回 nil, nil
func (e *engine) get(walletGlobalID uuid.UUID) (*actor, error) {
	if e == nil {
		return nil, errNilEngine
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	// Router 只计算下标, 实际 Actor 由对应 Shard 管理
	index, err := e.router.route(walletGlobalID)
	if err != nil {
		return nil, err
	}

	return e.shards[index].get(walletGlobalID)
}

// getOrCreate 根据钱包全局标识路由到 Shard, 查找或创建唯一 Actor
// 已有 Actor 保留运行时状态, 不使用传入的 State 覆盖余额和版本
func (e *engine) getOrCreate(state State) (*actor, error) {
	if e == nil {
		return nil, errNilEngine
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	index, err := e.router.route(state.WalletGlobalID)
	if err != nil {
		return nil, err
	}

	return e.shards[index].getOrCreate(state)
}

// submit 向指定钱包提交资金操作, 并等待 Actor 返回处理结果
func (e *engine) submit(ctx context.Context, walletGlobalID uuid.UUID, params RecordParams) (OperationResult, error) {
	if e == nil {
		return OperationResult{}, errNilEngine
	}
	if ctx == nil {
		return OperationResult{}, errNilEngineContext
	}
	if err := ctx.Err(); err != nil {
		return OperationResult{}, err
	}

	// 校验资金操作参数, 避免无效请求进入 Mailbox
	if err := params.validate(); err != nil {
		return OperationResult{}, err
	}

	// 读锁保护状态检查、路由和入队, 避免与 Engine 关闭交错
	e.mu.RLock()

	if e.closed {
		e.mu.RUnlock()
		return OperationResult{}, ErrEngineClosed
	}

	index, err := e.router.route(walletGlobalID)
	if err != nil {
		e.mu.RUnlock()
		return OperationResult{}, err
	}

	// 只接收已初始化的钱包, 不在资金提交时创建 Actor
	a, err := e.shards[index].get(walletGlobalID)
	if err != nil {
		e.mu.RUnlock()
		return OperationResult{}, err
	}
	if a == nil {
		e.mu.RUnlock()
		return OperationResult{}, ErrWalletActorNotFound
	}

	// Shard 下标同时也是目标 Kafka Partition
	message := newOperationMessage(params, int32(index))
	err = a.enqueue(ctx, message)

	// 入队完成后立即释放锁, 不持锁等待资金处理结果
	e.mu.RUnlock()

	if err != nil {
		return OperationResult{}, err
	}

	// 消息已由 Actor 接管, Context 只控制当前调用方的等待
	select {
	case reply := <-message.reply:
		return reply.result, reply.err

	case <-ctx.Done():
		return OperationResult{}, ctx.Err()
	}
}

// close 停止接收钱包请求, 并通知全部 Shard 关闭
func (e *engine) close() {
	if e == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return
	}

	// 持有写锁关闭全部 Shard, 避免与 getOrCreate 并发交错
	e.closed = true
	for _, s := range e.shards {
		s.close()
	}
}

// wait 等待全部 Shard 退出, 最后关闭共享 Worker Pool
// Context 只控制本次等待, 取消后可以再次调用 wait
func (e *engine) wait(ctx context.Context) error {
	if e == nil {
		return errNilEngine
	}
	if ctx == nil {
		return errNilEngineContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	e.mu.RLock()
	if !e.closed {
		e.mu.RUnlock()
		return ErrEngineNotClosed
	}
	e.mu.RUnlock()

	var waitErr error

	// 某个 Actor 出错也继续等待其他 Shard, 避免遗留后台任务
	for _, s := range e.shards {
		if err := s.wait(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			waitErr = errors.Join(waitErr, fmt.Errorf("wait wallet shard %d: %w", s.index, err))
		}
	}

	// Context 取消时保留 Worker Pool, 由下一次 wait 继续完成关闭
	if err := ctx.Err(); err != nil {
		return err
	}

	// 全部 Actor 退出后, 才允许关闭共享 Worker Pool
	if err := e.pool.close(); err != nil {
		waitErr = errors.Join(waitErr, fmt.Errorf("close wallet engine worker pool: %w", err))
	}

	return waitErr
}
