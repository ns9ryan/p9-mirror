package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"

	"github.com/redis/go-redis/v9"
)

const defaultRedisPrefix = "ipgeo"

// RedisCache Redis 驱动缓存。
type RedisCache struct {
	rdb    redis.UniversalClient // Redis 客户端
	prefix string                // 键前缀，实际键为 prefix:source|ip
}

// NewRedisCache 连接 Redis 并 Ping。
func NewRedisCache(cfg RedisConf) (*RedisCache, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, geo.ErrMissingRedis
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     strings.TrimSpace(cfg.Host),
		Password: cfg.Pass,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return &RedisCache{rdb: rdb, prefix: redisPrefix(cfg.KeyPrefix)}, nil
}

func redisPrefix(s string) string {
	s = strings.Trim(strings.TrimSpace(s), ":")
	if s == "" {
		return defaultRedisPrefix
	}
	return s
}

func (c *RedisCache) recKey(id string) string {
	return c.prefix + ":" + id
}

func (c *RedisCache) ipSetKey(ip string) string {
	return c.prefix + ":idx:" + ip
}

func (c *RedisCache) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

func (c *RedisCache) Get(ctx context.Context, id string) (*geo.Record, error) {
	key := c.recKey(id)
	pipe := c.rdb.Pipeline()
	get := pipe.Get(ctx, key)
	pttl := pipe.PTTL(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	raw, err := get.Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec, err := decodeRecord(raw)
	if err != nil || rec == nil {
		return rec, err
	}
	if ttl := pttl.Val(); ttl > 0 {
		rec.ExpiresAt = time.Now().Add(ttl)
	}
	return rec, nil
}

func (c *RedisCache) Set(ctx context.Context, rec *geo.Record) error {
	id := recordCacheID(rec)
	if rec == nil || id == "" {
		return nil
	}
	if rec.UpdatedAt.IsZero() {
		rec.UpdatedAt = time.Now()
	}
	raw, err := encodeRecord(rec)
	if err != nil {
		return err
	}
	key := c.recKey(id)
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, key, raw, expireDuration(rec.ExpiresAt))
	pipe.SAdd(ctx, c.ipSetKey(rec.IP), key)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *RedisCache) Touch(ctx context.Context, id string, expiresAt time.Time) error {
	if expiresAt.IsZero() {
		return nil
	}
	return c.rdb.ExpireAt(ctx, c.recKey(id), expiresAt).Err()
}

func (c *RedisCache) DeleteByIP(ctx context.Context, ip string) error {
	if ip == "" {
		return nil
	}
	idx := c.ipSetKey(ip)
	keys, err := c.rdb.SMembers(ctx, idx).Result()
	if err != nil {
		return err
	}
	del := append(keys, idx)
	return c.rdb.Del(ctx, del...).Err()
}

func expireDuration(at time.Time) time.Duration {
	if at.IsZero() {
		return 0
	}
	d := time.Until(at)
	if d < time.Millisecond {
		return time.Millisecond
	}
	return d
}
