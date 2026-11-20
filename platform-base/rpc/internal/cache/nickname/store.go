// Package nickname 实现低频变更的小型昵称池缓存。
package nickname

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
	"math/rand/v2"
	"oa.98ent.com/p9/platform-base/rpc/ent"
	entnickname "oa.98ent.com/p9/platform-base/rpc/ent/nickname"
	"oa.98ent.com/p9/platform-base/rpc/internal/cache"
	"time"
)

const revisionKey = "platform-base:nickname:revision"
const poolPrefix = "platform-base:nickname:pool:"

// Item 仅保存昵称业务字段，避免将 protobuf 运行时状态写入 MessagePack。
type Item struct {
	ID   int64
	Name string
}
type Store struct {
	ttl   time.Duration
	db    *ent.Client
	redis redis.UniversalClient
	group singleflight.Group
}

func New(db *ent.Client, r redis.UniversalClient, ttl time.Duration) *Store {
	return &Store{db: db, redis: r, ttl: ttl}
}
func (s *Store) read(ctx context.Context, key string) ([]*Item, bool) {
	b, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logx.WithContext(ctx).Errorf("nickname cache read: %v", err)
		}
		return nil, false
	}
	return s.decode(ctx, b)
}

func (s *Store) decode(ctx context.Context, payload string) ([]*Item, bool) {
	if payload == "" {
		return nil, false
	}
	// NF 表示数据库查询结果为空，命中时直接返回空列表。
	if payload == "NF" {
		return []*Item{}, true
	}
	var rows []*Item
	if err := msgpack.Unmarshal([]byte(payload), &rows); err != nil {
		logx.WithContext(ctx).Errorf("decode nickname cache: %v", err)
		return nil, false
	}
	for _, row := range rows {
		if row == nil || row.ID <= 0 || row.Name == "" {
			return nil, false
		}
	}
	return rows, true
}
func (s *Store) load(ctx context.Context) ([]*Item, error) {
	rows, err := s.db.Nickname.Query().Order(entnickname.ByID()).All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, &Item{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// All 按版本读取昵称池，数据库变更并递增版本后，旧回源任务只会回填旧键。
func (s *Store) All(ctx context.Context) ([]*Item, error) {
	if s.redis == nil {
		return s.load(ctx)
	}
	key, payload, err := cache.ReadVersioned(ctx, s.redis, revisionKey, poolPrefix)
	if err != nil {
		logx.WithContext(ctx).Errorf("nickname cache read: %v", err)
		return s.load(ctx)
	}
	if rows, ok := s.decode(ctx, payload); ok {
		return rows, nil
	}
	result := s.group.DoChan(key, func() (any, error) {
		// 共享回源任务使用独立超时，首个调用方取消时其他请求仍可等待结果。
		loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		// 共享回源前再次检查缓存，复用期间已写入的结果。
		if rows, ok := s.read(loadCtx, key); ok {
			return rows, nil
		}
		rows, err := s.load(loadCtx)
		if err != nil {
			return nil, err
		}
		ttl := s.ttl + time.Duration(rand.IntN(300))*time.Second
		var value any = "NF"
		if len(rows) == 0 {
			ttl = 30*time.Second + time.Duration(rand.IntN(10))*time.Second
		} else {
			b, e := msgpack.Marshal(rows)
			if e != nil {
				return rows, nil
			}
			value = b
		}
		if err = s.redis.Set(loadCtx, key, value, ttl).Err(); err != nil {
			logx.WithContext(loadCtx).Errorf("nickname cache write: %v", err)
		}
		return rows, nil
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-result:
		if r.Err != nil {
			return nil, r.Err
		}
		return r.Val.([]*Item), nil
	}
}

// Invalidate 在数据库提交后递增版本号，使用独立的超时上下文继续处理已提交的变更。
func (s *Store) Invalidate(ctx context.Context) {
	if s.redis == nil {
		return
	}
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	if err := s.redis.Incr(c, revisionKey).Err(); err != nil {
		logx.WithContext(c).Errorf("nickname cache invalidation failed (TTL bounds staleness): %v", err)
	}
}
