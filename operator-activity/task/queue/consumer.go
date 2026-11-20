// Package queue 实现基于 Redis 的异步消息队列消费者。
package queue

import (
	"context"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/utils"
)

// QueueConsumer Redis 队列消费者。
type QueueConsumer struct {
	ctx          context.Context
	redisQueue   *RedisQueue
	db           *ent.Client
	done         chan struct{}
	pollInterval time.Duration

	mu sync.RWMutex
	logx.Logger
}

// QueueStats 单个队列的统计信息。
type QueueStats struct {
	CumulativeProcessed int64
	LastPollTime        time.Time
	LastProcessTime     time.Time
}

func NewQueueConsumer(
	ctx context.Context,
	redisQueue *RedisQueue,
	db *ent.Client,
	pollInterval time.Duration,
) *QueueConsumer {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	return &QueueConsumer{
		ctx:          ctx,
		redisQueue:   redisQueue,
		db:           db,
		done:         make(chan struct{}),
		Logger:       logx.WithContext(ctx),
		pollInterval: pollInterval,
	}
}

func (qc *QueueConsumer) Start(ctx context.Context) error {
	qc.recoverProcessingQueues(ctx)
	go qc.consume(ctx)

	qc.Infof("Redis 队列消费者已启动, 轮询间隔=%v", qc.pollInterval)
	return nil
}

func (qc *QueueConsumer) consume(ctx context.Context) {
	// 创建 BatchProcessor
	instanceID := utils.GenerateInstanceID()
	processor := NewBatchProcessor(qc.ctx, qc.redisQueue, qc.db, instanceID)

	// 使用配置中的扫描间隔
	normalTicker := time.NewTicker(time.Duration(constants.NormalQueueScanInterval) * time.Second)
	retryTicker := time.NewTicker(time.Duration(constants.RetryQueueScanInterval) * time.Second)
	defer normalTicker.Stop()
	defer retryTicker.Stop()

	for {
		select {
		case <-normalTicker.C:
			// 处理正常队列（任务和派发）
			if err := processor.ProcessTaskBatchQueue(ctx); err != nil {
				qc.Errorf("处理任务队列失败: %v", err)
			}
			if err := processor.ProcessPayoutBatchQueue(ctx); err != nil {
				qc.Errorf("处理派发队列失败: %v", err)
			}

		case <-retryTicker.C:
			// 处理重试队列
			qc.Infof("扫描重试队列: task_retry_prefix=%s, payout_retry_prefix=%s",
				constants.RedisPrefixTaskRetry, constants.RedisPrefixPayoutRetry)
			if err := processor.ProcessRetryTaskBatchQueue(ctx); err != nil {
				qc.Errorf("处理任务重试队列失败: %v", err)
			}
			if err := processor.ProcessRetryPayoutBatchQueue(ctx); err != nil {
				qc.Errorf("处理派发重试队列失败: %v", err)
			}

		case <-qc.done:
			qc.Infof("Redis 队列消费者已停止")
			return

		case <-ctx.Done():
			qc.Infof("Redis 队列消费者上下文已取消")
			return
		}
	}
}

// FlushNow 关闭前刷新所有活动队列中的数据。
func (qc *QueueConsumer) FlushNow(ctx context.Context) error {
	qc.Infof("开始关闭前最后 Flush：刷新所有活动队列中的数据...")

	// 创建 BatchProcessor 用于最后的处理
	instanceID := utils.GenerateInstanceID()
	processor := NewBatchProcessor(qc.ctx, qc.redisQueue, qc.db, instanceID)

	// 处理所有待处理的队列
	if err := processor.ProcessTaskBatchQueue(ctx); err != nil {
		qc.Errorf("Flush 时处理任务队列失败: %v", err)
	}

	if err := processor.ProcessPayoutBatchQueue(ctx); err != nil {
		qc.Errorf("Flush 时处理派发队列失败: %v", err)
	}

	qc.Infof("关闭前最后 Flush 完成")
	return nil
}

func (qc *QueueConsumer) Stop() {
	close(qc.done)
	qc.Infof("Redis 队列消费者关闭中...")
}

// recoverProcessingQueues 启动时恢复处理中的队列
// 这里暂时不做任何操作，因为新架构中 Redis 中未删除的数据就是待处理的数据
func (qc *QueueConsumer) recoverProcessingQueues(ctx context.Context) {
	qc.Infof("启动时恢复处理中的队列...")
	// 新架构中，Redis 中尚未删除的数据就是待处理或待恢复的数据
	// 所以在启动时不需要特殊的恢复逻辑
}
