package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/types"
	"oa.98ent.com/p9/operator-activity/task/utils"
)

// AcquireBatchLock 尝试原子地获取批次处理锁 (SET NX PX)
// 返回: 是否成功获取锁，以及锁信息
func (rq *RedisQueue) AcquireBatchLock(ctx context.Context, batchDataKey string, instanceID string) (*types.BatchLockInfo, error) {
	lockKey := fmt.Sprintf("task:lock:%s", batchDataKey)

	// SET NX PX: 只有在 Key 不存在时才设置，并设置过期时间（毫秒）
	ttlMs := int64(constants.LockTTL * 1000) // 转换为毫秒
	result, err := rq.client.Set(ctx, lockKey, instanceID, time.Duration(ttlMs)*time.Millisecond).Result()

	if err != nil {
		if err == redis.Nil {
			rq.Debugf("未能获取批次锁 (lock_key=%s, instance_id=%s): 锁已被占用", lockKey, instanceID)
			return nil, fmt.Errorf("锁已被占用")
		}
		rq.Errorf("获取批次锁失败 (lock_key=%s): %v", lockKey, err)
		return nil, fmt.Errorf("获取锁失败: %w", err)
	}

	if result != "OK" {
		rq.Debugf("未能获取批次锁 (lock_key=%s): SET 返回非 OK", lockKey)
		return nil, fmt.Errorf("锁已被占用")
	}

	nowTime := time.Now()
	lockInfo := &types.BatchLockInfo{
		BatchID:    utils.ExtractBatchIDFromKey(batchDataKey),
		OwnerID:    instanceID,
		AcquiredAt: nowTime,
		ExpiresAt:  nowTime.Add(time.Duration(constants.LockTTL) * time.Second),
	}

	rq.Debugf("成功获取批次锁 (lock_key=%s, instance_id=%s, ttl=%ds)", lockKey, instanceID, constants.LockTTL)
	return lockInfo, nil
}

// RenewBatchLock 续期批次锁
// 只有当前 Owner 才能续期锁
// 使用 Lua 脚本保证原子性: 检查 Owner，然后续期
func (rq *RedisQueue) RenewBatchLock(ctx context.Context, batchDataKey string, instanceID string) error {
	lockKey := fmt.Sprintf("task:lock:%s", batchDataKey)

	// Lua 脚本：原子地检查 Owner 并续期
	renewScript := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)

	ttlMs := int64(constants.LockTTL * 1000) // 转换为毫秒
	result, err := renewScript.Run(ctx, rq.client, []string{lockKey}, instanceID, ttlMs).Result()

	if err != nil {
		rq.Errorf("续期批次锁失败 (lock_key=%s): %v", lockKey, err)
		return fmt.Errorf("续期锁失败: %w", err)
	}

	// 结果为 0 表示 Owner 不匹配或锁不存在
	if val, ok := result.(int64); ok && val == 0 {
		rq.Errorf("无法续期批次锁: Owner 校验失败或锁已过期 (lock_key=%s, instance_id=%s)", lockKey, instanceID)
		return fmt.Errorf("锁Owner不匹配或锁已过期")
	}

	rq.Debugf("成功续期批次锁 (lock_key=%s, instance_id=%s, ttl=%ds)", lockKey, instanceID, constants.LockTTL)
	return nil
}

// VerifyBatchLockOwner 校验批次锁的所有者
func (rq *RedisQueue) VerifyBatchLockOwner(ctx context.Context, batchDataKey string, instanceID string) (bool, error) {
	lockKey := fmt.Sprintf("task:lock:%s", batchDataKey)

	owner, err := rq.client.Get(ctx, lockKey).Result()
	if err != nil {
		if err == redis.Nil {
			rq.Debugf("批次锁已过期或不存在 (lock_key=%s)", lockKey)
			return false, nil
		}
		rq.Errorf("读取批次锁失败 (lock_key=%s): %v", lockKey, err)
		return false, fmt.Errorf("验证锁所有者失败: %w", err)
	}

	isOwner := owner == instanceID
	if !isOwner {
		rq.Errorf("批次锁 Owner 不匹配 (lock_key=%s, expected=%s, actual=%s)", lockKey, instanceID, owner)
	}

	return isOwner, nil
}

// ReleaseBatchLock 释放批次锁
// 只有当前 Owner 才能释放锁，使用 Lua 脚本保证原子性
func (rq *RedisQueue) ReleaseBatchLock(ctx context.Context, batchDataKey string, instanceID string) error {
	lockKey := fmt.Sprintf("task:lock:%s", batchDataKey)

	// Lua 脚本：原子地检查 Owner 然后删除锁
	releaseScript := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)

	result, err := releaseScript.Run(ctx, rq.client, []string{lockKey}, instanceID).Result()

	if err != nil {
		rq.Errorf("释放批次锁失败 (lock_key=%s): %v", lockKey, err)
		return fmt.Errorf("释放锁失败: %w", err)
	}

	// 结果为 0 表示 Owner 不匹配或锁不存在
	if val, ok := result.(int64); ok && val == 0 {
		rq.Errorf("无法释放批次锁: Owner 校验失败 (lock_key=%s, instance_id=%s)", lockKey, instanceID)
		return fmt.Errorf("锁Owner不匹配")
	}

	rq.Debugf("成功释放批次锁 (lock_key=%s, instance_id=%s)", lockKey, instanceID)
	return nil
}

// StartLockRenewalTicker 启动定时续期器
// 在处理过程中每隔 LockRenewalInterval 秒续期一次锁
// 返回: 一个 channel，用于接收续期中发生的错误
func (rq *RedisQueue) StartLockRenewalTicker(ctx context.Context, batchDataKey string, instanceID string) (chan error, func()) {
	errChan := make(chan error, 1)
	ticker := time.NewTicker(time.Duration(constants.LockRenewalInterval) * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := rq.RenewBatchLock(ctx, batchDataKey, instanceID); err != nil {
					rq.Errorf("批次锁续期失败: %v", err)
					errChan <- err
					return
				}
			}
		}
	}()

	// 返回一个停止函数
	stop := func() {
		ticker.Stop()
	}

	return errChan, stop
}
