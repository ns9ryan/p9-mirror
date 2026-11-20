package repository

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLockTTL      = 15 * time.Second
	defaultLockWait     = 3 * time.Second
	defaultLockInterval = 50 * time.Millisecond
	unlockScript        = `
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end`
)

// RedisUserPromoLocker 按活动+用户加分布式锁，串行化同一用户同一活动的发奖评估
type RedisUserPromoLocker struct {
	rdb      redis.Cmdable
	ttl      time.Duration
	wait     time.Duration // 自旋等待上限
	interval time.Duration // 重试间隔
}

func NewRedisUserPromoLocker(rdb redis.Cmdable) *RedisUserPromoLocker {
	return NewRedisUserPromoLockerWithTTL(rdb, defaultLockTTL)
}

func NewRedisUserPromoLockerWithTTL(rdb redis.Cmdable, ttl time.Duration) *RedisUserPromoLocker {
	if ttl <= 0 {
		ttl = defaultLockTTL
	}
	return &RedisUserPromoLocker{
		rdb:      rdb,
		ttl:      ttl,
		wait:     defaultLockWait,
		interval: defaultLockInterval,
	}
}

// WithWait 配置自旋等待上限与重试间隔（interval<=0 时用默认 50ms）
func (l *RedisUserPromoLocker) WithWait(wait, interval time.Duration) *RedisUserPromoLocker {
	if l == nil {
		return nil
	}
	if wait > 0 {
		l.wait = wait
	}
	if interval > 0 {
		l.interval = interval
	}
	return l
}

func lockKey(promoID, userID int64) string {
	return fmt.Sprintf("promo:lock:%d:%d", promoID, userID)
}

// Lock 自旋抢锁，成功返回 unlock；超时或 Redis 错误返回可重试错误
func (l *RedisUserPromoLocker) Lock(ctx context.Context, promoID, userID int64) (unlock func(), err error) {
	if l == nil || l.rdb == nil {
		return func() {}, nil
	}
	key := lockKey(promoID, userID)
	token := uuid.NewString()
	deadline := time.Now().Add(l.wait)
	if l.wait <= 0 {
		deadline = time.Now()
	}

	for {
		ok, err := l.rdb.SetNX(ctx, key, token, l.ttl).Result()
		if err != nil {
			return nil, perr.Retryable(fmt.Errorf("promo lock: %w", err))
		}
		if ok {
			return func() {
				_ = l.rdb.Eval(context.Background(), unlockScript, []string{key}, token).Err()
			}, nil
		}

		if time.Now().After(deadline) {
			return nil, perr.Retryablef("promo lock busy promo=%d user=%d", promoID, userID)
		}

		interval := l.interval
		if interval <= 0 {
			interval = defaultLockInterval
		}
		// 最后一次等待不超过剩余时间
		remain := time.Until(deadline)
		if remain < interval {
			interval = remain
		}
		if interval <= 0 {
			return nil, perr.Retryablef("promo lock busy promo=%d user=%d", promoID, userID)
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, perr.Retryable(fmt.Errorf("promo lock canceled promo=%d user=%d: %w", promoID, userID, ctx.Err()))
		case <-timer.C:
		}
	}
}
