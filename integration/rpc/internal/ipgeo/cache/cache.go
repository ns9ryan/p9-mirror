package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"oa.98ent.com/p9/integration/rpc/ent"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

// RedisConf Redis 缓存连接配置。
type RedisConf struct {
	Host      string // 地址，如 127.0.0.1:6379
	Pass      string // 密码，可空
	DB        int    // 库号
	KeyPrefix string // 键前缀，空则用 ipgeo
}

// 缓存驱动名常量。
const (
	CacheDriverTable  = "table"
	CacheDriverMemory = "memory"
	CacheDriverRedis  = "redis"
)

// Cache KV 缓存。key 为 name|规范化 IP。miss 时 Get 返回 nil, nil。
// 过期由 Lookup 按 Record.ExpiresAt 判断；表驱动 Touch 只改 expires_at；Redis 用 EXPIRE，Get 用 PTTL 回填。
type Cache interface {
	Get(ctx context.Context, key string) (*geo.Record, error)
	Set(ctx context.Context, rec *geo.Record) error
	Touch(ctx context.Context, key string, expiresAt time.Time) error
	DeleteByIP(ctx context.Context, ip string) error
}

// NormalizeCacheDriver 规范化驱动名，空则 table。
func NormalizeCacheDriver(driver string) string {
	d := strings.ToLower(strings.TrimSpace(driver))
	if d == "" {
		return CacheDriverTable
	}
	return d
}

// OpenCache 按驱动构造 Cache。
func OpenCache(driver string, db *ent.Client, redis RedisConf) (Cache, error) {
	switch NormalizeCacheDriver(driver) {
	case CacheDriverTable:
		if db == nil {
			return nil, geo.ErrMissingDSN
		}
		return NewTableCache(db), nil
	case CacheDriverMemory:
		return NewMemoryCache(), nil
	case CacheDriverRedis:
		return NewRedisCache(redis)
	default:
		return nil, fmt.Errorf("%w: %s", geo.ErrBadCacheDriver, driver)
	}
}

func encodeRecord(rec *geo.Record) ([]byte, error) {
	if rec == nil {
		return nil, nil
	}
	return json.Marshal(rec)
}

func decodeRecord(raw []byte) (*geo.Record, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	rec := &geo.Record{}
	if err := json.Unmarshal(raw, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// MemoryCache 进程内 KV 缓存。
type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]*geo.Record // key = source|ip
}

// NewMemoryCache 创建空的 MemoryCache。
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{data: make(map[string]*geo.Record)}
}

func (c *MemoryCache) Get(_ context.Context, key string) (*geo.Record, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rec, ok := c.data[key]
	if !ok {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}

func (c *MemoryCache) Set(_ context.Context, rec *geo.Record) error {
	if rec == nil || rec.IP == "" {
		return nil
	}
	id := recordCacheID(rec)
	if id == "" {
		return nil
	}
	cp := *rec
	if cp.UpdatedAt.IsZero() {
		cp.UpdatedAt = time.Now()
	}
	c.mu.Lock()
	c.data[id] = &cp
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Touch(_ context.Context, key string, expiresAt time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	rec, ok := c.data[key]
	if !ok {
		return nil
	}
	rec.ExpiresAt = expiresAt
	return nil
}

func (c *MemoryCache) DeleteByIP(_ context.Context, ip string) error {
	if ip == "" {
		return nil
	}
	suffix := "|" + ip
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.data {
		if key == ip || strings.HasSuffix(key, suffix) {
			delete(c.data, key)
		}
	}
	return nil
}

// Key 拼缓存键：source|ip。
func Key(source, ip string) string {
	if source == "" || ip == "" {
		return ""
	}
	return source + "|" + ip
}

func recordCacheID(rec *geo.Record) string {
	if rec == nil {
		return ""
	}
	return Key(rec.Source, rec.IP)
}

func parseCacheID(id string) (source, ip string) {
	source, ip, ok := strings.Cut(id, "|")
	if !ok {
		return "", id
	}
	return source, ip
}
