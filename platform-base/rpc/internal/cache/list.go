package cache

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

const maxListBytes = 4 << 20

// List 使用带版本号的 Redis 键缓存小型有序资源列表。
type List[T any] struct {
	ttl    time.Duration
	redis  redis.UniversalClient
	prefix string
	load   func(context.Context) ([]*T, error)
	group  singleflight.Group
}

func NewList[T any](client redis.UniversalClient, resource string, ttl time.Duration, load func(context.Context) ([]*T, error)) *List[T] {
	return &List[T]{redis: client, ttl: ttl, prefix: "platform-base:" + resource + ":", load: load}
}

func (c *List[T]) All(ctx context.Context) ([]*T, error) {
	if c.redis == nil {
		return c.load(ctx)
	}
	key, payload, err := ReadVersioned(ctx, c.redis, c.prefix+"revision", c.prefix+"list:msgpack:v1:")
	if err != nil {
		logx.WithContext(ctx).Errorf("read %s list cache: %v", c.prefix, err)
		return c.load(ctx)
	}
	if rows, ok := c.decode(ctx, payload); ok {
		return rows, nil
	}

	// 同一版本的未命中请求共享回源任务，任务使用独立超时。
	result := c.group.DoChan(key, func() (any, error) {
		loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		// 合并请求后再次读缓存，复用期间已写入的结果。
		if rows, ok := c.read(loadCtx, key); ok {
			return rows, nil
		}
		rows, err := c.load(loadCtx)
		if err != nil {
			return nil, err
		}
		payload, err := encodeList(rows)
		if err == nil {
			// 随机抖动分散过期时间，空列表采用较短的 TTL。
			ttl := c.ttl + time.Duration(rand.IntN(300))*time.Second
			if len(rows) == 0 {
				ttl = 30*time.Second + time.Duration(rand.IntN(10))*time.Second
			}
			err = c.redis.Set(loadCtx, key, payload, ttl).Err()
		}
		if err != nil {
			logx.WithContext(loadCtx).Errorf("write %s list cache: %v", c.prefix, err)
		}
		// 回填失败时仍返回已从数据库读取的数据。
		return rows, nil
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-result:
		if r.Err != nil {
			return nil, r.Err
		}
		return r.Val.([]*T), nil
	}
}

func (c *List[T]) read(ctx context.Context, key string) ([]*T, bool) {
	payload, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logx.WithContext(ctx).Errorf("read %s list cache: %v", c.prefix, err)
		}
		return nil, false
	}
	return c.decode(ctx, payload)
}

func (c *List[T]) decode(ctx context.Context, payload string) ([]*T, bool) {
	if payload == "" {
		return nil, false
	}
	var rows []*T
	if err := decodeList([]byte(payload), &rows); err != nil {
		logx.WithContext(ctx).Errorf("decode %s list cache: %v", c.prefix, err)
		return nil, false
	}
	return rows, true
}

// Invalidate 在数据库写入成功后递增版本号。并发回源若持有旧版本，只会写入旧键。
func (c *List[T]) Invalidate(ctx context.Context) {
	if c.redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	if err := c.redis.Incr(ctx, c.prefix+"revision").Err(); err != nil {
		logx.WithContext(ctx).Errorf("invalidate %s list cache: %v", c.prefix, err)
	}
}

func encodeList[T any](rows []*T) ([]byte, error) {
	payload, err := msgpack.Marshal(rows)
	if err != nil {
		return nil, err
	}
	if len(payload) > maxListBytes {
		return nil, fmt.Errorf("list exceeds %d bytes", maxListBytes)
	}
	return payload, nil
}

func decodeList[T any](payload []byte, rows *[]*T) error {
	if len(payload) > maxListBytes {
		return fmt.Errorf("list exceeds %d bytes", maxListBytes)
	}
	if err := msgpack.Unmarshal(payload, rows); err != nil {
		return err
	}

	for _, row := range *rows {
		if row == nil {
			return errors.New("cached list contains a null row")
		}
	}
	return nil
}
