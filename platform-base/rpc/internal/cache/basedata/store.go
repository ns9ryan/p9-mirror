// Package basedata 配置 platform-base 基础资源列表缓存。
package basedata

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"oa.98ent.com/p9/platform-base/rpc/ent"
	"oa.98ent.com/p9/platform-base/rpc/ent/currency"
	"oa.98ent.com/p9/platform-base/rpc/ent/defaultavatar"
	"oa.98ent.com/p9/platform-base/rpc/ent/defaultavatarframe"
	"oa.98ent.com/p9/platform-base/rpc/ent/region"
	"oa.98ent.com/p9/platform-base/rpc/ent/timezone"
	"oa.98ent.com/p9/platform-base/rpc/ent/vipdefaulticongroup"
	"oa.98ent.com/p9/platform-base/rpc/internal/cache"
)

type Store struct {
	Timezone            *cache.List[ent.Timezone]
	Currency            *cache.List[ent.Currency]
	Region              *cache.List[ent.Region]
	VipDefaultIconGroup *cache.List[ent.VipDefaultIconGroup]
	DefaultAvatarFrame  *cache.List[ent.DefaultAvatarFrame]
	DefaultAvatar       *cache.List[ent.DefaultAvatar]
}

func New(db *ent.Client, client redis.UniversalClient, ttl time.Duration) *Store {
	return &Store{
		Timezone: cache.NewList(client, "timezone", ttl, func(ctx context.Context) ([]*ent.Timezone, error) {
			return db.Timezone.Query().Order(timezone.BySortNo(), timezone.ByID()).All(ctx)
		}),
		Currency: cache.NewList(client, "currency", ttl, func(ctx context.Context) ([]*ent.Currency, error) {
			return db.Currency.Query().Order(currency.BySortNo(), currency.ByID()).All(ctx)
		}),
		Region: cache.NewList(client, "region", ttl, func(ctx context.Context) ([]*ent.Region, error) {
			return db.Region.Query().Order(region.BySortNo(), region.ByID()).All(ctx)
		}),
		VipDefaultIconGroup: cache.NewList(client, "vip-default-icon", ttl, func(ctx context.Context) ([]*ent.VipDefaultIconGroup, error) {
			return db.VipDefaultIconGroup.Query().Order(vipdefaulticongroup.ByMinLevel()).All(ctx)
		}),
		DefaultAvatarFrame: cache.NewList(client, "default-avatar-frame", ttl, func(ctx context.Context) ([]*ent.DefaultAvatarFrame, error) {
			return db.DefaultAvatarFrame.Query().Order(defaultavatarframe.ByID()).All(ctx)
		}),
		DefaultAvatar: cache.NewList(client, "default-avatar", ttl, func(ctx context.Context) ([]*ent.DefaultAvatar, error) {
			return db.DefaultAvatar.Query().Order(defaultavatar.ByID()).All(ctx)
		}),
	}
}
