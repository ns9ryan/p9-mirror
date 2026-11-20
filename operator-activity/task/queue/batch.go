package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/types"
	"oa.98ent.com/p9/operator-activity/task/utils"
)

// BatchAggregator 批次聚合器
// 简化设计：每个时间窗口对应一个 Redis Key
// 不限制消息数量，直接按时间窗口分组
type BatchAggregator struct {
	logger logx.Logger
}

// NewBatchAggregator 创建新的批次聚合器
func NewBatchAggregator(logger logx.Logger) *BatchAggregator {
	return &BatchAggregator{
		logger: logger,
	}
}

// ========== 批次写入相关函数 ==========

// 任务事件写入批次
func (rq *RedisQueue) DispatchTaskEventToBatch(ctx context.Context, event *types.Event) error {
	batchPrefix := constants.RedisPrefixTask
	return rq.WriteBatchToRedis(ctx, batchPrefix, event)
}

// 派发奖励事件写入批次
func (rq *RedisQueue) DispatchPayoutEventToBatch(ctx context.Context, event *types.Event) error {
	batchPrefix := constants.RedisPrefixPayout
	return rq.WriteBatchToRedis(ctx, batchPrefix, event)
}

// WriteBatchToRedis 将事件数据写入 Redis Hash
// 使用 HSet 原子地写入单个字段
// 对于重试事件，优先使用 OccurredAt 保证同一事件的所有重试都在同一批次
// 返回: 错误信息
func (rq *RedisQueue) WriteBatchToRedis(ctx context.Context, batchPrefix string, event *types.Event) error {
	if event == nil {
		return fmt.Errorf("没有要写入的事件")
	}

	// 计算时间窗口：优先使用事件的原始发生时间（OccurredAt）
	// 这样重试事件可以和原事件归到同一批次，避免生成过多的小批次
	var timeWindow string
	if !event.OccurredAt.IsZero() {
		timeWindow = event.OccurredAt.Format(constants.BatchTimeWindowFormat)
	} else {
		timeWindow = time.Now().Format(constants.BatchTimeWindowFormat)
	}

	batchDataKey := rq.GetBatchDataKey(batchPrefix, timeWindow)

	// 生成 Hash Field
	field := utils.FieldKey(event, event.PeriodKey)

	// 序列化事件为 JSON
	data, err := json.Marshal(event)
	if err != nil {
		rq.Errorf("序列化事件失败: field=%s, err=%v", field, err)
		return fmt.Errorf("事件序列化失败: %w", err)
	}

	// 写入到 Redis Hash
	if err := rq.client.HSet(ctx, batchDataKey, field, string(data)).Err(); err != nil {
		rq.Errorf("写入批次失败: key=%s, field=%s, err=%v", batchDataKey, field, err)
		return fmt.Errorf("写入批次失败: %w", err)
	}

	rq.Debugf("批次数据写入成功: key=%s, field=%s, retryCount=%d", batchDataKey, field, event.RetryCount)
	return nil
}

// DeleteBatchFromRedis 删除已处理的批次 Key
// 处理完成后调用此方法清理数据
func (rq *RedisQueue) DeleteBatchFromRedis(ctx context.Context, batchDataKey string) error {
	result, err := rq.client.Del(ctx, batchDataKey).Result()
	if err != nil {
		rq.Errorf("删除批次 Key 失败: key=%s, err=%v", batchDataKey, err)
		return fmt.Errorf("删除批次失败: %w", err)
	}

	if result == 0 {
		rq.Errorf("批次 Key 不存在或已删除: key=%s", batchDataKey)
	} else {
		rq.Debugf("批次 Key 已删除: key=%s", batchDataKey)
	}

	return nil
}

// ReadBatchFromRedis 从 Redis 读取整个批次的所有事件
// 返回: 事件列表、错误信息
func (rq *RedisQueue) ReadBatchFromRedis(ctx context.Context, batchDataKey string) ([]*types.Event, error) {
	result, err := rq.client.HGetAll(ctx, batchDataKey).Result()
	if err != nil {
		rq.Errorf("读取批次数据失败: key=%s, err=%v", batchDataKey, err)
		return nil, fmt.Errorf("读取批次失败: %w", err)
	}

	if len(result) == 0 {
		rq.Errorf("批次 Key 为空或不存在: key=%s", batchDataKey)
		return nil, fmt.Errorf("批次Key为空")
	}

	var events []*types.Event
	for field, data := range result {
		var event types.Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			rq.Errorf("反序列化事件失败: field=%s, err=%v", field, err)
			continue
		}
		events = append(events, &event)
	}

	rq.Debugf("批次读取成功: key=%s, 事件数=%d", batchDataKey, len(events))
	return events, nil
}

// DeleteSuccessfulFieldsFromBatch 从批次中删除成功处理的字段，保留失败的数据用于重试
// successfulFields: 成功处理的字段名列表
// 返回: 删除的字段数
func (rq *RedisQueue) DeleteSuccessfulFieldsFromBatch(ctx context.Context, batchDataKey string, successfulFields []string) (int, error) {
	if len(successfulFields) == 0 {
		return 0, nil
	}

	result, err := rq.client.HDel(ctx, batchDataKey, successfulFields...).Result()
	if err != nil {
		rq.Errorf("删除成功字段失败: key=%s, err=%v", batchDataKey, err)
		return 0, fmt.Errorf("删除成功字段失败: %w", err)
	}

	rq.Infof("已从批次中删除成功字段: key=%s, 删除数=%d (保留失败数据用于重试)", batchDataKey, result)
	return int(result), nil
}

// DeleteFailedFieldsFromBatch 已弃用：改为使用 DeleteSuccessfulFieldsFromBatch
// 保留此函数以兼容旧代码
func (rq *RedisQueue) DeleteFailedFieldsFromBatch(ctx context.Context, batchDataKey string, failedFields []string) (int, error) {
	if len(failedFields) == 0 {
		return 0, nil
	}

	result, err := rq.client.HDel(ctx, batchDataKey, failedFields...).Result()
	if err != nil {
		rq.Errorf("删除失败字段失败: key=%s, err=%v", batchDataKey, err)
		return 0, fmt.Errorf("删除失败字段失败: %w", err)
	}

	rq.Infof("已从批次中删除失败字段: key=%s, 删除数=%d", batchDataKey, result)
	return int(result), nil
}

// GetBatchSize 获取批次中的字段总数
func (rq *RedisQueue) GetBatchSize(ctx context.Context, batchDataKey string) (int64, error) {
	result, err := rq.client.HLen(ctx, batchDataKey).Result()
	if err != nil {
		rq.Errorf("获取批次大小失败: key=%s, err=%v", batchDataKey, err)
		return 0, fmt.Errorf("获取批次大小失败: %w", err)
	}
	return result, nil
}

// ScanBatchKeys 扫描符合前缀的所有批次 Key
// 使用 SCAN 命令避免阻塞
// 返回: 批次 Key 列表
func (rq *RedisQueue) ScanBatchKeys(ctx context.Context, batchPrefix string) ([]string, error) {
	var keys []string
	pattern := fmt.Sprintf("%s:%s:*", batchPrefix, rq.serviceName)

	iter := rq.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		rq.Errorf("扫描批次 Key 失败: pattern=%s, err=%v", pattern, err)
		return nil, fmt.Errorf("扫描批次Key失败: %w", err)
	}

	rq.Debugf("扫描得到批次 Key: pattern=%s, 数量=%d", pattern, len(keys))
	return keys, nil
}
