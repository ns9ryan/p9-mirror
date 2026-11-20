package queue

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/membertaskreward"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/reward"
	"oa.98ent.com/p9/operator-activity/task/types"
	"oa.98ent.com/p9/operator-activity/task/utils"
)

// BatchProcessor 批次处理器
// 负责扫描、认领、处理和清理批次
type BatchProcessor struct {
	redisQueue *RedisQueue
	db         *ent.Client
	instanceID string
	logx.Logger
}

// NewBatchProcessor 创建新的批次处理器
func NewBatchProcessor(ctx context.Context, redisQueue *RedisQueue, db *ent.Client, instanceID string) *BatchProcessor {
	return &BatchProcessor{
		redisQueue: redisQueue,
		db:         db,
		instanceID: instanceID,
		Logger:     logx.WithContext(ctx),
	}
}

// ProcessBatchQueue 处理单个批次队列
// 1. 尝试原子认领（SET NX PX）
// 2. 如果认领成功，读取数据并处理
// 3. 启动 Lock 续期
// 4. 处理完成后删除 Key 并释放 Lock
// 5. 如果处理失败，只修改失败数据，保留数据继续重试
func (bp *BatchProcessor) ProcessBatchQueue(ctx context.Context, batchDataKey string, batchType string, processFunc func(context.Context, []*types.Event) []*types.Event) error {
	bp.Debugf("[批处理器] 开始处理批次队列: key=%s, type=%s", batchDataKey, batchType)

	// 记录活跃批次数
	ActiveBatches.WithLabelValues(batchType).Inc()
	defer ActiveBatches.WithLabelValues(batchType).Dec()

	// 步骤 1：尝试原子认领锁
	_, err := bp.redisQueue.AcquireBatchLock(ctx, batchDataKey, bp.instanceID)
	if err != nil {
		bp.Debugf("[批处理器] 获取锁失败: key=%s, err=%v", batchDataKey, err)
		return fmt.Errorf("获取锁失败: %w", err)
	}

	bp.Debugf("[批处理器] 锁已获取: key=%s, owner=%s", batchDataKey, bp.instanceID)

	// 步骤 2：创建上下文和取消函数，用于 Lock 续期
	lockCtx, lockCancel := context.WithCancel(ctx)
	defer lockCancel()

	// 步骤 3：启动 Lock 续期
	errChan, stopRenewal := bp.redisQueue.StartLockRenewalTicker(lockCtx, batchDataKey, bp.instanceID)
	defer stopRenewal()

	// 步骤 4：读取批次数据
	events, err := bp.redisQueue.ReadBatchFromRedis(ctx, batchDataKey)
	if err != nil {
		bp.Errorf("[批处理器] 读取批次失败: key=%s, err=%v", batchDataKey, err)
		// 读取失败，释放锁并返回错误
		_ = bp.redisQueue.ReleaseBatchLock(ctx, batchDataKey, bp.instanceID)
		return fmt.Errorf("读取批次失败: %w", err)
	}

	bp.Infof("[批处理器] 批次数据已加载: key=%s, event_count=%d, trace_ids=%s", batchDataKey, len(events), func() string {
		// 用逗号分隔拼接traceID
		traceIDs := make([]string, len(events))
		for i, event := range events {
			traceIDs[i] = event.TraceID
		}
		return strings.Join(traceIDs, ",")
	}())

	// 步骤 5：处理事件
	startTime := time.Now()
	failedEvents := processFunc(ctx, events)
	duration := time.Since(startTime)

	bp.Debugf("[批处理器] 批次处理完成: key=%s, total=%d, failed=%d, duration=%v",
		batchDataKey, len(events), len(failedEvents), duration)

	// 记录批次处理指标（只记录实际发生的结果）
	successCount := len(events) - len(failedEvents)
	if successCount > 0 {
		// 只在有成功事件时才记录 success
		RecordBatchProcessing(batchType, duration, "success", successCount)
	}
	if len(failedEvents) > 0 {
		// 只在有失败事件时才记录 failed
		RecordBatchProcessing(batchType, duration, "failed", len(failedEvents))
	}

	// 步骤 6：处理失败数据
	if len(failedEvents) > 0 {
		bp.Errorf("[批处理器] 处理部分失败: key=%s, failed_count=%d", batchDataKey, len(failedEvents))

		// 推送失败的事件到重试队列
		retryPrefix := bp.getRetryPrefix(batchType)
		if retryPrefix != "" {
			for _, event := range failedEvents {
				event.RetryCount++
				if err := bp.redisQueue.WriteBatchToRedis(ctx, retryPrefix, event); err != nil {
					bp.Errorf("[批处理器] 写入重试队列失败: key=%s, event=%+v, err=%v",
						batchDataKey, event, err)
				}
			}
		}
	}

	// 步骤 7：删除整个批次 Key
	// 判断是否为重试队列，然后决定是否删除
	isRetryQueue := batchType == constants.RedisPrefixTaskRetry || batchType == constants.RedisPrefixPayoutRetry
	if !isRetryQueue || len(failedEvents) == 0 {
		// 情况1：普通队列 → 删除（失败事件已推送到重试队列）
		// 情况2：重试队列且无失败事件 → 删除
		//   - 要么所有事件都成功处理
		//   - 要么所有失败事件都已超出重试次数被标记为拒绝
		err = bp.redisQueue.DeleteBatchFromRedis(ctx, batchDataKey)
		if err != nil {
			bp.Errorf("[批处理器] 删除批次失败: key=%s, err=%v", batchDataKey, err)
		}
	} else {
		// 情况3：重试队列且有仍需重试的失败事件 → 保留
		// 失败事件已推回重试队列（retry_count++），等待下一个扫描周期重新处理
		bp.Debugf("[批处理器] 重试队列保留失败事件以便下次处理: key=%s, failed_count=%d", batchDataKey, len(failedEvents))
	}

	// 步骤 8：释放锁
	err = bp.redisQueue.ReleaseBatchLock(ctx, batchDataKey, bp.instanceID)
	if err != nil {
		bp.Errorf("[批处理器] 释放锁失败: key=%s, err=%v", batchDataKey, err)
	}

	// 检查 Lock 续期过程中是否发生错误
	select {
	case err := <-errChan:
		bp.Errorf("[批处理器] 锁续期错误: %v", err)
		return err
	default:
	}

	bp.Debugf("[批处理器] 批次处理完毕: key=%s, success_count=%d",
		batchDataKey, len(events)-len(failedEvents))

	return nil
}

// getRetryPrefix 根据当前批次类型，获取对应的重试队列前缀
// 普通队列 → 重试队列；重试队列 → 自己（保留在重试队列中继续处理）
func (bp *BatchProcessor) getRetryPrefix(currentPrefix string) string {
	switch currentPrefix {
	case constants.RedisPrefixTask:
		return constants.RedisPrefixTaskRetry
	case constants.RedisPrefixPayout:
		return constants.RedisPrefixPayoutRetry
	case constants.RedisPrefixTaskRetry:
		// 已在任务重试队列，继续推回重试队列（retry_count会递增）
		// 直到 processRetryTaskEvents 中达到 MaxRetry 限制才标记拒绝
		return constants.RedisPrefixTaskRetry
	case constants.RedisPrefixPayoutRetry:
		// 已在派发重试队列，继续推回重试队列（retry_count会递增）
		// 直到 processRetryPayoutEvents 中达到 MaxRetry 限制才标记拒绝
		return constants.RedisPrefixPayoutRetry
	default:
		return ""
	}
}

// ScanAndProcessBatchQueue 扫描前缀中的所有批次并处理（并发模式）
// 1. SCAN 扫描符合前缀的所有 Key
// 2. 使用 goroutine 并发处理多个批次（限制最大并发数）
// 3. 返回处理结果摘要
func (bp *BatchProcessor) ScanAndProcessBatchQueue(ctx context.Context, prefix string, processFunc func(context.Context, []*types.Event) []*types.Event) (*types.BatchResult, error) {
	// 扫描所有符合前缀的 Key
	keys, err := bp.redisQueue.ScanBatchKeys(ctx, prefix)
	if err != nil {
		bp.Errorf("[批处理器] 扫描批次键失败: prefix=%s, err=%v", prefix, err)
		return nil, err
	}

	if len(keys) == 0 {
		bp.Debugf("[批处理器] 未发现批次键: prefix=%s", prefix)
		return &types.BatchResult{
			Type:     prefix,
			Total:    0,
			Success:  0,
			Failed:   0,
			Duration: 0,
		}, nil
	}

	bp.Infof("[批处理器] 发现批次键: prefix=%s, count=%d (开始并发处理)", prefix, len(keys))

	// 处理每个批次
	result := &types.BatchResult{
		Type:    prefix,
		Total:   len(keys),
		Success: 0,
		Failed:  0,
	}

	startTime := time.Now()

	// 使用信号量限制并发数（最多5个并发处理）
	maxConcurrency := 5
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, key := range keys {
		wg.Add(1)
		go func(batchKey string) {
			defer wg.Done()
			// 获取信号量令牌
			sem <- struct{}{}
			defer func() { <-sem }()

			// 尝试处理批次
			if err := bp.ProcessBatchQueue(ctx, batchKey, prefix, processFunc); err != nil {
				bp.Errorf("[批处理器] 处理批次失败: key=%s, err=%v", batchKey, err)
				mu.Lock()
				result.Failed++
				mu.Unlock()
			} else {
				mu.Lock()
				result.Success++
				mu.Unlock()
			}
		}(key)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	result.Duration = time.Since(startTime)

	bp.Infof("[批处理器] 并发处理完成: type=%s, total=%d, success=%d, failed=%d, duration=%v",
		prefix, result.Total, result.Success, result.Failed, result.Duration)

	return result, nil
}

// ProcessTaskBatchQueue 处理任务批次队列的包装方法
func (bp *BatchProcessor) ProcessTaskBatchQueue(ctx context.Context) error {
	return bp.ScanAndProcessBatchQueueSimple(ctx, constants.RedisPrefixTask, bp.processTaskEvents)
}

// ProcessPayoutBatchQueue 处理派发批次队列的包装方法
func (bp *BatchProcessor) ProcessPayoutBatchQueue(ctx context.Context) error {
	return bp.ScanAndProcessBatchQueueSimple(ctx, constants.RedisPrefixPayout, bp.processPayoutEvents)
}

// ProcessRetryTaskBatchQueue 处理任务重试批次队列的包装方法
func (bp *BatchProcessor) ProcessRetryTaskBatchQueue(ctx context.Context) error {
	return bp.ScanAndProcessBatchQueueSimple(ctx, constants.RedisPrefixTaskRetry, bp.processRetryTaskEvents)
}

// ProcessRetryPayoutBatchQueue 处理派发重试批次队列的包装方法
func (bp *BatchProcessor) ProcessRetryPayoutBatchQueue(ctx context.Context) error {
	return bp.ScanAndProcessBatchQueueSimple(ctx, constants.RedisPrefixPayoutRetry, bp.processRetryPayoutEvents)
}

// ScanAndProcessBatchQueueSimple 简化版的扫描和处理方法
func (bp *BatchProcessor) ScanAndProcessBatchQueueSimple(ctx context.Context, prefix string, processFunc func(context.Context, []*types.Event) []*types.Event) error {
	_, err := bp.ScanAndProcessBatchQueue(ctx, prefix, processFunc)
	return err
}

// ========== 事件处理函数（这些需要与现有的业务逻辑集成）==========

// processTaskEvents 处理任务事件
// 为每个事件创建 MemberTaskReward 记录
// 如果配置为自动派发，生成派发任务到派发队列
// 返回处理失败的事件列表
func (bp *BatchProcessor) processTaskEvents(ctx context.Context, events []*types.Event) []*types.Event {
	var failedEvents []*types.Event
	successCount := 0

	for _, event := range events {
		if err := bp.processOneTaskEvent(ctx, event); err != nil {
			bp.Errorf("[批处理器] 处理任务事件失败: kafka_id=%s, member_id=%d, traceID=%s, err=%v",
				event.KafkaID, event.MemberID, event.TraceID, err)
			failedEvents = append(failedEvents, event)
		} else {
			successCount++
		}
	}

	bp.Infof("[批处理器] 任务事件处理完毕: total=%d, success=%d, failed=%d",
		len(events), successCount, len(failedEvents))

	return failedEvents
}

// processPayoutEvents 处理派发事件
// 直接从事件中获取派发订单信息，调用钱包服务完成派发
// 返回处理失败的事件列表
func (bp *BatchProcessor) processPayoutEvents(ctx context.Context, events []*types.Event) []*types.Event {
	var failedEvents []*types.Event

	// 获取钱包客户端
	walletClient := reward.NewGrpcWalletClient(ctx)

	for _, event := range events {
		// 直接从事件中构建派发任务
		payoutTask := types.PayoutTask{
			OrderNo:      event.OrderNo,
			OperatorCode: event.OperatorCode,
			MemberID:     event.MemberID,
			Amount:       event.Amount,
			TaskCode:     event.TaskCode,
			RetryCount:   event.RetryCount,
			MaxRetry:     event.MaxRetry,
		}

		bp.Debugf("[批处理器] 开始派发: order_no=%s, member_id=%d, amount=%d, task_code=%s, traceID=%s",
			payoutTask.OrderNo, payoutTask.MemberID, payoutTask.Amount, payoutTask.TaskCode, event.TraceID)

		if err := bp.processOnePayoutTask(ctx, walletClient, payoutTask, event.TraceID); err != nil {
			bp.Errorf("[批处理器] 处理派发事件失败: order_no=%s, traceID=%s, err=%v",
				payoutTask.OrderNo, event.TraceID, err)
			failedEvents = append(failedEvents, event)
		} else {
			bp.Debugf("[批处理器] 派发成功: order_no=%s, traceID=%s, 状态已更新为99",
				payoutTask.OrderNo, event.TraceID)
		}
	}

	bp.Infof("[批处理器] 派发事件处理完毕: total=%d, success=%d, failed=%d",
		len(events), len(events)-len(failedEvents), len(failedEvents))

	return failedEvents
}

// processRetryTaskEvents 处理重试任务事件
// 与普通任务处理相同，但检查是否已超过最大重试次数
// 超过限制的事件会被标记为拒绝状态，不进入下一轮重试
// 返回处理失败且仍在重试范围内的事件列表
func (bp *BatchProcessor) processRetryTaskEvents(ctx context.Context, events []*types.Event) []*types.Event {
	bp.Debugf("[批处理器] 处理 %d 个重试任务事件", len(events))

	var failedEvents []*types.Event
	successCount := 0

	for _, event := range events {
		// 检查是否已超过最大重试次数
		if event.RetryCount >= event.MaxRetry {
			bp.Errorf("[批处理器] 重试任务事件超过最大重试次数: kafka_id=%s, member_id=%d, retry_count=%d, max_retry=%d",
				event.KafkaID, event.MemberID, event.RetryCount, event.MaxRetry)
			// 打印详细event
			bp.Errorf("[批处理器] 详细事件信息: %+v", event)
			status := constants.TaskStatusRejected
			taskConfig, err := bp.redisQueue.GetTaskConfig(ctx, event.OperatorCode, event.TaskCode)
			if err != nil {
				bp.Errorf("获取任务配置失败: operator_code=%s, task_code=%s, err=%v", event.OperatorCode, event.TaskCode, err)
				continue
			}
			periodKey := utils.CalculatePeriodKey(taskConfig, event.OccurredAt)
			// 先尝试更新
			affectedRows, err := bp.db.MemberTaskReward.Update().
				Where(
					membertaskreward.OperatorCodeEQ(event.OperatorCode),
					membertaskreward.MemberIDEQ(event.MemberID),
					membertaskreward.TaskCodeEQ(event.TaskCode),
					membertaskreward.PeriodKeyEQ(periodKey),
					membertaskreward.StatusNotIn(constants.TaskStatusDispatched),
				).
				SetStatus(status).
				Save(ctx)

			if err != nil {
				bp.Errorf("更新任务进度状态失败: operator_code=%s, member_id=%d, task_code=%s, period_key=%s, err=%v",
					event.OperatorCode, event.MemberID, event.TaskCode, periodKey, err)
				continue
			}

			// 如果没有更新任何记录，则创建新记录
			if affectedRows == 0 {
				orderNo := utils.RandomStr()
				_, err := bp.db.MemberTaskReward.Create().
					SetKafkaID(event.KafkaID).
					SetOperatorCode(event.OperatorCode).
					SetOrderNo(orderNo).
					SetMemberID(event.MemberID).
					SetCategory(taskConfig.Category).
					SetPeriodKey(event.PeriodKey).
					SetTaskCode(taskConfig.TaskCode).
					SetStatus(status).
					SetPayoutMode(taskConfig.PayoutMode).
					SetAmount(taskConfig.RewardAmount).
					Save(ctx)

				if err != nil {
					bp.Errorf("创建任务进度记录失败: kafka_id=%s, task_code=%s, err=%v", event.KafkaID, event.TaskCode, err)
					continue
				}

				bp.Infof("创建任务进度记录: kafka_id=%s, task_code=%s, status=%d",
					event.KafkaID, event.TaskCode, status)
				continue
			}

			bp.Infof("更新任务进度状态成功: kafka_id=%s, member_id=%d, task_code=%s, status=%d",
				event.KafkaID, event.MemberID, event.TaskCode, status)
			continue
		}

		// 尝试处理任务事件
		if err := bp.processOneTaskEvent(ctx, event); err != nil {
			bp.Errorf("[批处理器] 处理重试任务事件失败: kafka_id=%s, member_id=%d, retry_count=%d, err=%v",
				event.KafkaID, event.MemberID, event.RetryCount, err)
			failedEvents = append(failedEvents, event)
		} else {
			successCount++
		}
	}

	bp.Infof("[批处理器] 重试任务事件处理完毕: total=%d, success=%d, failed=%d",
		len(events), successCount, len(failedEvents))

	return failedEvents
}

// processRetryPayoutEvents 处理重试派发事件
// 直接从事件中获取派发订单信息，检查重试次数后调用钱包服务
// 超过限制的派发订单会被标记为派发拒绝状态
// 返回处理失败且仍在重试范围内的事件列表
func (bp *BatchProcessor) processRetryPayoutEvents(ctx context.Context, events []*types.Event) []*types.Event {
	bp.Debugf("[批处理器] 处理 %d 个重试派发事件", len(events))

	var failedEvents []*types.Event

	// 获取钱包客户端
	walletClient := reward.NewGrpcWalletClient(ctx)

	for _, event := range events {
		// 检查是否已超过最大重试次数
		if event.RetryCount >= event.MaxRetry {
			bp.Errorf("[批处理器] 重试派发事件超过最大重试次数: order_no=%s, retry_count=%d, max_retry=%d",
				event.OrderNo, event.RetryCount, event.MaxRetry)
			// 打印详细event
			bp.Errorf("[批处理器] 详细事件信息: %+v", event)
			// 更新订单状态为派发拒绝，打印更新数量
			if affectedRows, err := bp.db.MemberTaskReward.Update().
				Where(
					membertaskreward.OrderNoEQ(event.OrderNo),
					membertaskreward.StatusNotIn(constants.TaskStatusDispatched),
				).
				SetStatus(constants.TaskStatusDispatchRejected).
				Save(ctx); err != nil {
				bp.Errorf("[批处理器] 更新订单状态为拒绝失败: order_no=%s, err=%v",
					event.OrderNo, err)
			} else {
				bp.Infof("[批处理器] 更新订单状态为拒绝成功: order_no=%s, affected_rows=%d",
					event.OrderNo, affectedRows)
			}
			// 标记为拒绝状态，不作为失败事件返回（因为已无法恢复）
			continue
		}

		taskCode := constants.EventTypeToTaskCode[event.EventType]

		// 直接从事件中构建派发任务
		payoutTask := types.PayoutTask{
			OrderNo:      event.OrderNo,
			OperatorCode: event.OperatorCode,
			MemberID:     event.MemberID,
			Amount:       event.Amount,
			TaskCode:     taskCode,
			RetryCount:   event.RetryCount,
			MaxRetry:     event.MaxRetry,
		}

		// 尝试派发
		if err := bp.processOnePayoutTask(ctx, walletClient, payoutTask, event.TraceID); err != nil {
			bp.Errorf("[批处理器] 处理重试派发事件失败: order_no=%s, retry_count=%d, traceID=%s, err=%v",
				payoutTask.OrderNo, payoutTask.RetryCount, event.TraceID, err)
			failedEvents = append(failedEvents, event)
		}
	}

	bp.Infof("[批处理器] 重试派发事件处理完毕: total=%d, success=%d, failed=%d",
		len(events), len(events)-len(failedEvents), len(failedEvents))
	return failedEvents
}

// ========== 事件处理辅助函数 ==========

// processOneTaskEvent 处理单个任务事件
// 创建 MemberTaskReward 记录，如果自动派发则生成派发任务到派发队列
func (bp *BatchProcessor) processOneTaskEvent(ctx context.Context, event *types.Event) error {
	startTime := time.Now()
	bp.Debugf("[批处理器] 处理任务事件: kafka_id=%s, member_id=%d, period_key=%s, traceID=%s",
		event.KafkaID, event.MemberID, event.PeriodKey, event.TraceID)

	// 获取任务配置
	taskConfig, err := bp.redisQueue.GetTaskConfig(ctx, event.OperatorCode, event.TaskCode)
	if err != nil {
		bp.Errorf("[批处理器] 获取任务配置失败: err=%v", err)
		RecordEventProcessing("task", time.Since(startTime), "failed")
		return err
	}

	// 生成随机订单号
	orderNo := utils.RandomStr()
	if orderNo == "" {
		RecordEventProcessing("task", time.Since(startTime), "failed")
		return fmt.Errorf("生成订单号失败")
	}
	// // 模拟失败
	// if event.EventType == constants.EventTypeMemberRegistered {
	// 	bp.Errorf("[批处理器] 模拟任务事件处理失败: order_no=%s, traceID=%s", orderNo, event.TraceID)
	// 	RecordEventProcessing("task", time.Since(startTime), "failed")
	// 	return fmt.Errorf("模拟任务事件处理失败")
	// }

	// 确定初始状态
	initialStatus := constants.TaskStatusPendingDispatch
	if taskConfig.PayoutMode == constants.PayoutModeManual {
		initialStatus = constants.TaskStatusPendingClaim
	}

	// 创建奖励订单
	rewardOrder, err := bp.db.MemberTaskReward.Create().
		SetKafkaID(event.KafkaID).
		SetOperatorCode(event.OperatorCode).
		SetOrderNo(orderNo).
		SetMemberID(event.MemberID).
		SetCategory(taskConfig.Category).
		SetPeriodKey(event.PeriodKey).
		SetTaskCode(taskConfig.TaskCode).
		SetStatus(initialStatus).
		SetPayoutMode(taskConfig.PayoutMode).
		SetAmount(taskConfig.RewardAmount).
		Save(ctx)
	if err != nil {
		// 如果是主键冲突，可能是重复事件，直接忽略
		if ent.IsConstraintError(err) {
			bp.Infof("[批处理器] 重复事件，周期键已存在: period_key=%s, member_id=%d, kafka_id=%s",
				event.PeriodKey, event.MemberID, event.KafkaID)
			RecordEventProcessing("task", time.Since(startTime), "success") // 重复不算失败
			RecordDuplicatePeriodKey("task")                                // 记录重复数据
			return nil
		}
		bp.Errorf("[批处理器] 创建奖励订单失败: kafka_id=%s, err=%v", event.KafkaID, err)
		RecordEventProcessing("task", time.Since(startTime), "failed")
		return fmt.Errorf("创建奖励订单失败: %w", err)
	}

	// 记录实际入库的数据
	RecordStoredSuccessfully("task")
	RecordEventProcessing("task", time.Since(startTime), "success")

	bp.Debugf("[批处理器] 奖励订单已创建: kafka_id=%s, trace_id=%s, order_no=%s, status=%d",
		event.KafkaID, event.TraceID, orderNo, initialStatus)

	// 如果状态为待派发，则生成派发事件到派发队列
	if rewardOrder.Status == constants.TaskStatusPendingDispatch {
		// 创建派发事件
		payoutEvent := &types.Event{
			EventType:    event.EventType, // 保留原事件类型用于重试处理
			OperatorCode: rewardOrder.OperatorCode,
			MemberID:     rewardOrder.MemberID,
			TaskCode:     rewardOrder.TaskCode,
			Amount:       rewardOrder.Amount,
			OrderNo:      rewardOrder.OrderNo, // 派发时需要的订单号
			PeriodKey:    event.PeriodKey,     // 继承原事件的周期键
			KafkaID:      event.KafkaID,       // 追踪源消息
			TraceID:      event.TraceID,       // 继承链路追踪 ID
			OccurredAt:   time.Now(),
			RetryCount:   0,
			MaxRetry:     constants.MaxPayoutRetries,
		}

		// 写入派发队列
		if err := bp.redisQueue.DispatchPayoutEventToBatch(ctx, payoutEvent); err != nil {
			bp.Errorf("[批处理器] 写入派发事件到redis失败: order_no=%s, err=%v", orderNo, err)
			return fmt.Errorf("写入派发事件失败: %w", err)
		}

		bp.Debugf("[批处理器] 派发事件已写入redis: order_no=%s, task_code=%s",
			orderNo, taskConfig.TaskCode)
	}

	return nil
}

// processOnePayoutTask 处理单个派发任务
// 调用钱包服务完成派发，成功则更新订单状态
func (bp *BatchProcessor) processOnePayoutTask(ctx context.Context, walletClient types.WalletClient, task types.PayoutTask, traceID string) error {
	startTime := time.Now()
	bp.Debugf("[批处理器] 处理派发任务: order_no=%s, member_id=%d, amount=%d, traceID=%s",
		task.OrderNo, task.MemberID, task.Amount, traceID)

	// 调用钱包服务进行派发
	payoutReq := &types.PayoutRequest{
		OpCode:          task.OperatorCode,
		MemberID:        task.MemberID,
		WalletOrderNo:   task.OrderNo,
		BusinessType:    "TASK_REWARD",
		BusinessOrderID: task.OrderNo,
		Currency:        "USD",
		AmountMinor:     task.Amount,
		Remark:          fmt.Sprintf("Task reward for task_code=%s", task.TaskCode),
	}

	payoutResp, err := walletClient.Payout(ctx, payoutReq)
	if err != nil {
		bp.Errorf("[批处理器] 派发RPC调用失败: order_no=%s, traceID=%s, err=%v", task.OrderNo, traceID, err)
		RecordEventProcessing("payout", time.Since(startTime), "failed")
		RecordPayoutAttempt("failed", task.RetryCount+1)
		return fmt.Errorf("派发RPC调用失败: %w", err)
	}

	// // 模拟失败
	// if payoutResp != nil {
	// 	bp.Errorf("[批处理器] 模拟任务事件处理失败: order_no=%s, traceID=%s", task.OrderNo, traceID)
	// 	RecordEventProcessing("task", time.Since(startTime), "failed")
	// 	return fmt.Errorf("模拟任务事件处理失败")
	// }

	// 检查派发结果
	if !payoutResp.Success {
		bp.Errorf("[批处理器] 派发失败: order_no=%s, traceID=%s, code=%s, message=%s",
			task.OrderNo, traceID, payoutResp.Code, payoutResp.Message)
		RecordEventProcessing("payout", time.Since(startTime), "failed")
		RecordPayoutAttempt("failed", task.RetryCount+1)
		return fmt.Errorf("派发失败: code=%s, message=%s", payoutResp.Code, payoutResp.Message)
	}

	bp.Debugf("[批处理器] 派发成功: order_no=%s, traceID=%s, transaction_id=%s",
		task.OrderNo, traceID, payoutResp.WalletTransactionID)

	// 派发成功，更新订单状态为已派发
	if err := bp.db.MemberTaskReward.Update().
		Where(
			membertaskreward.OrderNoEQ(task.OrderNo),
		).
		SetStatus(constants.TaskStatusDispatched).
		SetWalletID(payoutResp.WalletID).
		SetWalletTransactionID(payoutResp.WalletTransactionID).
		Exec(ctx); err != nil {
		bp.Errorf("[批处理器] 更新订单状态失败: order_no=%s, traceID=%s, err=%v", task.OrderNo, traceID, err)
		RecordEventProcessing("payout", time.Since(startTime), "failed")
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	bp.Debugf("[批处理器] 订单状态已更新为已派发: order_no=%s, traceID=%s", task.OrderNo, traceID)
	RecordStoredSuccessfully("payout") // 记录派发成功（实际入库）
	RecordEventProcessing("payout", time.Since(startTime), "success")
	RecordPayoutAttempt("success", task.RetryCount+1)
	return nil
}
