package consumer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	errEmptyBrokers           = errors.New("wallet consumer: empty brokers")
	errEmptyTopic             = errors.New("wallet consumer: empty topic")
	errEmptyPartitions        = errors.New("wallet consumer: empty partitions")
	errNilContext             = errors.New("wallet consumer: nil context")
	errNilHandler             = errors.New("wallet consumer: nil handler")
	errConsumerAlreadyStarted = errors.New("wallet consumer: already started")
	errConsumerClosed         = errors.New("wallet consumer: already closed")
)

// Config 配置钱包资金事件消费者
type Config struct {
	Brokers    []string             // Kafka Broker 地址
	Topic      string               // 资金事件 Topic
	Partitions map[int32]kgo.Offset // 已校验的 Partition 起始 Offset
}

// Handler 处理单条钱包资金事件
type Handler func(context.Context, *kgo.Record) error

// Consumer 负责顺序消费钱包资金事件
type Consumer struct {
	client *kgo.Client // Kafka 消费客户端

	mu      sync.Mutex
	cancel  context.CancelFunc // 消费循环取消函数
	done    chan struct{}      // 消费循环退出信号
	started bool               // 是否已经启动
	closed  bool               // 是否已经关闭
	runErr  error              // 消费循环运行错误

	closeOnce sync.Once // 确保客户端只关闭一次
}

// New 创建固定 Partition 的钱包资金事件消费者, Partitions 必须由 Checkpoint 和 Kafka 日志范围校验后提供
func New(config Config) (*Consumer, error) {
	if len(config.Brokers) == 0 {
		return nil, errEmptyBrokers
	}

	for _, broker := range config.Brokers {
		if broker == "" || strings.TrimSpace(broker) != broker {
			return nil, errors.New("wallet consumer: invalid broker address")
		}
	}

	if config.Topic == "" || strings.TrimSpace(config.Topic) != config.Topic {
		return nil, errEmptyTopic
	}

	if len(config.Partitions) == 0 {
		return nil, errEmptyPartitions
	}

	// 复制分区配置, 避免外部修改原 Map
	partitions := make(map[int32]kgo.Offset, len(config.Partitions))
	for partition, offset := range config.Partitions {
		if partition < 0 {
			return nil, fmt.Errorf("%w: partition=%d", errInvalidConsumeOffsets, partition)
		}

		partitions[partition] = offset
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			config.Topic: partitions,
		}),
		kgo.ConsumeResetOffset(kgo.NoResetOffset()), // Offset 越界时禁止自动重置
	)
	if err != nil {
		return nil, fmt.Errorf("create wallet consumer: %w", err)
	}

	return &Consumer{
		client: client,
	}, nil
}

// Start 启动后台消费循环
func (c *Consumer) Start(ctx context.Context, handler Handler) error {
	if c == nil {
		return errConsumerClosed
	}
	if ctx == nil {
		return errNilContext
	}
	if handler == nil {
		return errNilHandler
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errConsumerClosed
	}
	if c.started {
		return errConsumerAlreadyStarted
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)

	c.cancel = cancel
	c.done = make(chan struct{})
	c.started = true

	go c.run(runCtx, handler)

	return nil
}

// run 执行 Kafka Poll 消费循环
func (c *Consumer) run(ctx context.Context, handler Handler) {
	var runErr error

	defer func() {
		c.mu.Lock()
		c.runErr = runErr
		close(c.done)
		c.mu.Unlock()
	}()

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		fetches := c.client.PollFetches(ctx)

		if ctx.Err() != nil {
			return
		}

		// 拉取失败时停止消费, 避免静默跳过资金事件
		if fetchErrors := fetches.Errors(); len(fetchErrors) > 0 {
			runErr = fmt.Errorf(
				"poll wallet events: topic=%s partition=%d: %w",
				fetchErrors[0].Topic,
				fetchErrors[0].Partition,
				fetchErrors[0].Err,
			)
			return
		}

		// 顺序处理资金事件, Handler 失败后停止消费
		var handlerErr error

		fetches.EachRecord(func(record *kgo.Record) {
			if handlerErr != nil {
				return
			}

			// 服务停止时不再处理后续消息
			if ctx.Err() != nil {
				handlerErr = ctx.Err()
				return
			}

			if err := handler(ctx, record); err != nil {
				handlerErr = fmt.Errorf(
					"handle wallet event: topic=%s partition=%d offset=%d: %w",
					record.Topic,
					record.Partition,
					record.Offset,
					err,
				)
			}
		})

		// Handler 失败时保留原始错误, 即使此时服务正在停止
		if handlerErr != nil {
			runErr = handlerErr
			return
		}
	}
}

// Stop 停止消费循环并等待其退出
func (c *Consumer) Stop(ctx context.Context) error {
	if c == nil {
		return nil
	}
	if ctx == nil {
		return errNilContext
	}

	c.mu.Lock()

	if !c.started {
		c.mu.Unlock()
		return nil
	}

	cancel := c.cancel
	done := c.done

	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	select {
	case <-done:
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.runErr

	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close 关闭 Kafka 消费客户端, 应在 Stop 成功后调用
func (c *Consumer) Close() {
	if c == nil {
		return
	}

	c.mu.Lock()
	c.closed = true

	if c.cancel != nil {
		c.cancel()
	}

	c.mu.Unlock()

	c.closeOnce.Do(func() {
		if c.client != nil {
			c.client.Close()
		}
	})
}
