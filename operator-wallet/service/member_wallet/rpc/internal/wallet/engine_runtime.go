package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EngineConfig 配置钱包运行时
type EngineConfig struct {
	// Kafka 配置
	KafkaBrokers []string // Kafka Broker 地址
	KafkaTopic   string   // 资金事件 Topic

	// 并发配置
	ShardCount      int // Shard 数量, 必须与 Topic Partition 数量一致
	WorkerCount     int // 共享 Worker 数量
	MailboxCapacity int // 每个 Actor 的 Mailbox 容量

	PublishTimeout time.Duration // 单次 Kafka 写入超时时间
}

// runtimeProducer 定义钱包运行时所需的 Kafka 生产者能力
type runtimeProducer interface {
	eventProducer
	close()
}

// Engine 对外提供钱包运行时能力, 内部管理 Actor、Worker Pool 和 Kafka Producer
type Engine struct {
	core *engine // 钱包 Engine 核心
	// producer *kafkaProducer // Kafka 资金事件生产者
	producer runtimeProducer // Kafka 资金事件生产者

	producerClose sync.Once // 确保 Producer 只关闭一次
}

// NewEngine 创建钱包运行时及 Kafka Producer
// ctx 必须是钱包服务生命周期 Context, 不能使用单次 RPC 请求 Context
// Engine 关闭并等待所有 Actor 退出后, 才能取消该 Context
func NewEngine(ctx context.Context, config EngineConfig) (*Engine, error) {
	if ctx == nil {
		return nil, errNilEngineContext
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("create wallet engine: %w", err)
	}

	// 创建 Kafka Producer
	producer, err := newKafkaProducer(config.KafkaBrokers, config.KafkaTopic)
	if err != nil {
		return nil, fmt.Errorf("create wallet engine producer: %w", err)
	}

	// 创建资金操作 Handler
	handler, err := newOperationHandler(ctx, producer, config.PublishTimeout)
	if err != nil {
		producer.close()
		return nil, fmt.Errorf("create wallet engine handler: %w", err)
	}

	// 创建 Engine 核心及全部 Shard
	core, err := newEngine(
		config.ShardCount,      // Shard 数量, 必须与 Topic Partition 数量一致
		config.WorkerCount,     // 共享 Worker 数量
		config.MailboxCapacity, // 每个 Actor 的 Mailbox 容量
		handler,
	)
	if err != nil {
		producer.close()
		return nil, fmt.Errorf("create wallet engine core: %w", err)
	}

	return &Engine{
		core:     core,
		producer: producer,
	}, nil
}

// HasWallet 检查 Engine 是否已经管理指定钱包, 已故障的 Actor 仍视为已注册, 不允许重新创建
func (e *Engine) HasWallet(walletGlobalID uuid.UUID) (bool, error) {
	if e == nil || e.core == nil {
		return false, errNilEngine
	}

	a, err := e.core.get(walletGlobalID)
	if err != nil {
		return false, err
	}

	return a != nil, nil
}

// RegisterWallet 注册钱包 Actor, 已存在时复用原 Actor, 不覆盖其运行时状态
func (e *Engine) RegisterWallet(state State) error {
	if e == nil || e.core == nil {
		return errNilEngine
	}

	_, err := e.core.getOrCreate(state)
	if err != nil {
		return fmt.Errorf("register wallet actor: %w", err)
	}

	return nil
}

// Submit 提交一笔钱包资金操作
func (e *Engine) Submit(ctx context.Context, walletGlobalID uuid.UUID, params RecordParams) (OperationResult, error) {
	if e == nil || e.core == nil {
		return OperationResult{}, errNilEngine
	}

	return e.core.submit(ctx, walletGlobalID, params)
}

// StopAccepting 停止接收新的资金操作, 已接收的操作继续处理
func (e *Engine) StopAccepting() {
	if e == nil || e.core == nil {
		return
	}

	e.core.close()
}

// Close 关闭钱包运行时, 等待 Actor 和 Worker 完全退出后再关闭 Producer
func (e *Engine) Close(ctx context.Context) error {
	if e == nil || e.core == nil {
		return nil
	}
	if ctx == nil {
		return errNilEngineContext
	}

	// 先停止接收新请求, 并关闭全部 Actor
	e.StopAccepting()

	// 等待已接受的资金操作处理完成
	waitErr := e.core.wait(ctx)

	// 等待超时或 Worker 尚未退出时, 不能提前关闭 Producer
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(waitErr, ErrWorkerPoolBusy) {
		return waitErr
	}

	// Actor 全部退出后才释放 Kafka Producer
	e.producerClose.Do(func() {
		if e.producer != nil {
			e.producer.close()
		}
	})

	return waitErr
}

// GetBalance 查询指定钱包当前运行时余额
func (e *Engine) GetBalance(ctx context.Context, walletGlobalID uuid.UUID) (BalanceSnapshot, error) {
	if e == nil || e.core == nil {
		return BalanceSnapshot{}, errNilEngine
	}

	return e.core.getBalance(ctx, walletGlobalID)
}
