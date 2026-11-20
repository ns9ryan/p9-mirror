package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	fieldCount  = "count"
	fieldAmount = "amount"
)

// RedisDepositCounter 窗口期充值 Redis 累计计数器
type RedisDepositCounter struct {
	rdb redis.Cmdable
}

func NewRedisDepositCounter(rdb redis.Cmdable) *RedisDepositCounter {
	return &RedisDepositCounter{rdb: rdb}
}

// IncrOrLoad key 不存在则 load 回源并写入；已存在则对 orderID 幂等累加 amount
func (c *RedisDepositCounter) IncrOrLoad(
	ctx context.Context,
	key string,
	orderID int64,
	amount int64,
	ttl time.Duration,
	load func(context.Context) (count int, sum int64, err error),
) (int, int64, error) {
	if c == nil || c.rdb == nil {
		return load(ctx)
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	orderSetKey := key + ":oids"
	added, err := c.rdb.SAdd(ctx, orderSetKey, strconv.FormatInt(orderID, 10)).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis sadd order: %w", err)
	}
	_ = c.rdb.Expire(ctx, orderSetKey, ttl).Err()

	if added == 0 {
		// 同一订单已计入，直接读
		return c.get(ctx, key, load)
	}

	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis exists: %w", err)
	}

	if exists == 0 {
		count, sum, err := load(ctx)
		if err != nil {
			_, _ = c.rdb.SRem(ctx, orderSetKey, strconv.FormatInt(orderID, 10)).Result()
			return 0, 0, err
		}
		// HSetNX count：抢到初始化权则写入 DB 全量（已含当前单）
		ok, err := c.rdb.HSetNX(ctx, key, fieldCount, count).Result()
		if err != nil {
			return 0, 0, fmt.Errorf("redis hsetnx: %w", err)
		}
		if ok {
			if err := c.rdb.HSet(ctx, key, fieldAmount, sum).Err(); err != nil {
				return 0, 0, fmt.Errorf("redis hset amount: %w", err)
			}
			_ = c.rdb.Expire(ctx, key, ttl).Err()
			return count, sum, nil
		}
		// 未抢到：他人已初始化，对本单做增量
	}

	count, err := c.rdb.HIncrBy(ctx, key, fieldCount, 1).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis hincr count: %w", err)
	}
	sum, err := c.rdb.HIncrBy(ctx, key, fieldAmount, amount).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis hincr amount: %w", err)
	}
	_ = c.rdb.Expire(ctx, key, ttl).Err()
	return int(count), sum, nil
}

func (c *RedisDepositCounter) get(
	ctx context.Context,
	key string,
	load func(context.Context) (int, int64, error),
) (int, int64, error) {
	vals, err := c.rdb.HMGet(ctx, key, fieldCount, fieldAmount).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis hmget: %w", err)
	}
	if len(vals) < 2 || vals[0] == nil || vals[1] == nil {
		return load(ctx)
	}
	count, err := strconv.Atoi(fmt.Sprint(vals[0]))
	if err != nil {
		return 0, 0, err
	}
	sum, err := strconv.ParseInt(fmt.Sprint(vals[1]), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return count, sum, nil
}
