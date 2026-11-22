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
	errEmptyGroupID           = errors.New("wallet consumer: empty group ID")
	errNilContext             = errors.New("wallet consumer: nil context")
	errNilHandler             = errors.New("wallet consumer: nil handler")
	errConsumerAlreadyStarted = errors.New("wallet consumer: already started")
	errConsumerClosed         = errors.New("wallet consumer: already closed")
)

// Config 配置钱包资金事件消费者
type Config struct {
	Brokers []string // Kafka Broker 地址
	Topic   string   // 资金事件 Topic
	GroupID string   // Consumer Group ID
}

// Handler 处理单条钱包资金事件
type Handler func(context.Context, *kgo.Record) error

// Consumer 负责消费钱包资金事件
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

// New 创建钱包资金事件消费者
func New(config Config) (*Consumer, error) {
	if len(config.Brokers) == 0 {
		return nil, errEmptyBrokers
	}

	for _, broker := range config.Brokers {
		if broker == "" || strings.TrimSpace(broker) != broker {
			return nil, fmt.Errorf("wallet consumer: invalid broker address")
		}
	}

	if config.Topic == "" || strings.TrimSpace(config.Topic) != config.Topic {
		return nil, errEmptyTopic
	}

	if config.GroupID == "" || strings.TrimSpace(config.GroupID) != config.GroupID {
		return nil, errEmptyGroupID
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumerGroup(config.GroupID),
		kgo.ConsumeTopics(config.Topic),
		kgo.DisableAutoCommit(), // 由数据库事务成功后手动提交 Offset
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

		// 当前阶段遇到分区拉取错误时停止消费, 避免静默跳过
		if fetchErrors := fetches.Errors(); len(fetchErrors) > 0 {
			runErr = fmt.Errorf(
				"poll wallet events: topic=%s partition=%d: %w",
				fetchErrors[0].Topic,
				fetchErrors[0].Partition,
				fetchErrors[0].Err,
			)
			return
		}

		// 当前先串行处理拉取到的记录, 后续再实现分区级并行
		var handlerErr error

		fetches.EachRecord(func(record *kgo.Record) {
			if handlerErr != nil {
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
