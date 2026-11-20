package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/taskconfig"
	"oa.98ent.com/p9/operator-activity/task/constants"
)

// BatchInfo 批次元数据。
type BatchInfo struct {
	BatchID  string `json:"batch_id"`
	TaskCode string `json:"task_code"`
}

// RedisQueue Redis 消息队列（基于 Hash 结构）。
// 同一 operator/member/period 的事件会覆盖，达到周期内聚合去重效果。
type RedisQueue struct {
	client      *redis.Client
	db          *ent.Client
	serviceName string
	logx.Logger
}

func NewRedisQueue(ctx context.Context, redisClient *redis.Client, db *ent.Client, serviceName string) *RedisQueue {
	return &RedisQueue{
		client:      redisClient,
		db:          db,
		Logger:      logx.WithContext(ctx),
		serviceName: serviceName,
	}
}

func (rq *RedisQueue) GetClient() *redis.Client {
	return rq.client
}

// 获取taskconfig redis缓存，如果不存在从db中获取
func (rq *RedisQueue) GetTaskConfig(ctx context.Context, operatorCode string, taskCode string) (*ent.TaskConfig, error) {
	if operatorCode == "" || taskCode == "" {
		return nil, fmt.Errorf("operatorCode和taskCode不能为空: operatorCode=%s, taskCode=%s", operatorCode, taskCode)
	}
	// 从redis缓存获取taskConfig，如果存在则直接返回
	result, err := rq.client.Get(ctx, fmt.Sprintf("%s:task_config:%s:%s", rq.serviceName, operatorCode, taskCode)).Result()
	if err == nil {
		// 将 result 反序列化为 *ent.TaskConfig 并返回
		taskConfig := &ent.TaskConfig{}
		err := json.Unmarshal([]byte(result), taskConfig)
		if err == nil {
			return taskConfig, nil
		}
	}
	logx.Infof("从 Redis 获取 task_config 失败，准备从 DB 获取: operatorCode=%s, taskCode=%s", operatorCode, taskCode)
	// 如果从 Redis 获取失败，继续从 DB 获取
	taskConfig, err := rq.db.TaskConfig.Query().Where(
		taskconfig.OperatorCodeEQ(operatorCode),
		taskconfig.TaskCodeEQ(taskCode),
		taskconfig.Status(constants.TaskConfigStatusEnabled),
	).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询任务配置失败: %w", err)
	}
	// 将从 DB 获取到的 taskConfig 缓存到 Redis
	data, err := json.Marshal(taskConfig)
	if err == nil {
		rq.client.Set(ctx, fmt.Sprintf("%s:task_config:%s:%s", rq.serviceName, operatorCode, taskCode), data, 0) // 不设置过期
	}
	return taskConfig, nil
}

// InvalidateTaskConfigCache 清除指定 task_config 的 Redis 缓存
// 在配置更新后调用此方法，确保下次查询时从数据库重新加载最新配置
func (rq *RedisQueue) InvalidateTaskConfigCache(ctx context.Context, operatorCode string, taskCode string) error {
	if operatorCode == "" || taskCode == "" {
		return fmt.Errorf("operatorCode和taskCode不能为空: operatorCode=%s, taskCode=%s", operatorCode, taskCode)
	}

	cacheKey := fmt.Sprintf("%s:task_config:%s:%s", rq.serviceName, operatorCode, taskCode)
	result := rq.client.Del(ctx, cacheKey)

	if result.Err() != nil {
		rq.Errorf("[缓存更新] 清除 task_config 缓存失败: key=%s, err=%v", cacheKey, result.Err())
		return result.Err()
	}

	deletedCount := result.Val()
	if deletedCount > 0 {
		rq.Infof("[缓存更新] task_config 缓存已清除: operatorCode=%s, taskCode=%s", operatorCode, taskCode)
	}

	return nil
}

// taskHashTag 让同一个 taskCode 的 queue、processing、batch Key 位于 Redis Cluster 同一个 Slot。
// taskCode 不应包含花括号；如允许任意字符，建议在业务入口限制或编码 taskCode。
func (rq *RedisQueue) taskHashTag(taskCode string) string {
	return "{" + taskCode + "}"
}

func (rq *RedisQueue) queueKey(prefix, taskCode string) string {
	return fmt.Sprintf("%s:%s:queue", prefix, rq.taskHashTag(taskCode))
}

func (rq *RedisQueue) GetTaskQueueKeyPrefix(operatorCode string) string {
	return fmt.Sprintf("%s:%s:%s", constants.RedisPrefixTask, rq.serviceName, operatorCode)
}

func (rq *RedisQueue) GetPayoutQueueKeyPrefix(operatorCode string) string {
	return fmt.Sprintf("%s:%s:%s", constants.RedisPrefixPayout, rq.serviceName, operatorCode)
}

func (rq *RedisQueue) GetTaskRetryQueueKeyPrefix(operatorCode string) string {
	return fmt.Sprintf("%s:%s:%s", constants.RedisPrefixTaskRetry, rq.serviceName, operatorCode)
}

func (rq *RedisQueue) GetPayoutRetryQueueKeyPrefix(operatorCode string) string {
	return fmt.Sprintf("%s:%s:%s", constants.RedisPrefixPayoutRetry, rq.serviceName, operatorCode)
}

// GetTaskRetryQueueKey 获取任务重试队列的完整 key
func (rq *RedisQueue) GetTaskRetryQueueKey(operatorCode, taskCode string) string {
	return rq.queueKey(rq.GetTaskRetryQueueKeyPrefix(operatorCode), taskCode)
}

// GetPayoutRetryQueueKey 获取派发重试队列的完整 key
func (rq *RedisQueue) GetPayoutRetryQueueKey(operatorCode, taskCode string) string {
	return rq.queueKey(rq.GetPayoutRetryQueueKeyPrefix(operatorCode), taskCode)
}

// ========== 新批次系统相关的 Key 生成方法 ==========

// GetBatchDataKey 获取批次数据存储的 Redis Hash Key
// 简化设计：格式 prefix:serviceName:timeWindow（不使用序号）
func (rq *RedisQueue) GetBatchDataKey(prefix string, batchID string) string {
	return fmt.Sprintf("%s:%s:%s", prefix, rq.serviceName, batchID)
}

// GetBatchLockKey 获取批次锁的 Redis Key
// 格式: task:lock:prefix:serviceName:timeWindow
func (rq *RedisQueue) GetBatchLockKey(prefix string, batchID string) string {
	return fmt.Sprintf("task:lock:%s:%s:%s", prefix, rq.serviceName, batchID)
}

// GetBatchTypePrefix 根据类型获取对应的前缀
// types: "task", "payout", "taskretry", "payoutretry"
func GetBatchTypePrefix(batchType string) string {
	switch batchType {
	case "task":
		return constants.RedisPrefixTask
	case "payout":
		return constants.RedisPrefixPayout
	case "taskretry":
		return constants.RedisPrefixTaskRetry
	case "payoutretry":
		return constants.RedisPrefixPayoutRetry
	default:
		return batchType
	}
}
