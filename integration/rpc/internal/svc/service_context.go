package svc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/config"
	"oa.98ent.com/p9/integration/rpc/internal/db"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/cache"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
	"oa.98ent.com/p9/integration/rpc/service"

	"github.com/zeromicro/go-zero/core/logx"
)

// ServiceContext RPC 依赖注入容器（go-zero svc）。
//
// 组装关系：
//   - Deps：跨域聚合（Deps.IpGeo = IP/GEO 管理 CRUD）
//   - IpGeo：查询引擎 Lookup（仅 Search），与 Deps.IpGeo 共用 Client/Cache
//
// 当前全服务共用一套 PostgreSQL；后续多业务在 Deps 上增加域服务即可。
type ServiceContext struct {
	Config config.Config // 配置
	Deps   *service.Deps // 业务层依赖（按域：IpGeo / …）
	IpGeo  *ipgeo.Lookup // IP/GEO 查询编排
}

// NewServiceContext 打开库与缓存、种子 Provider、构造 Deps 与 Lookup。
func NewServiceContext(c config.Config) *ServiceContext {
	if d := strings.TrimSpace(c.DB.Driver); d != "" && d != "postgres" {
		logx.Must(fmt.Errorf("unsupported db.driver %q, want postgres", d))
	}
	client, err := db.Open(c.DB.DSN)
	logx.Must(err)

	cch, err := cache.OpenCache(c.IpGeo.CacheDriver, client, cache.RedisConf{
		Host:      c.IpGeo.CacheRedis.Host,
		Pass:      c.IpGeo.CacheRedis.Pass,
		DB:        c.IpGeo.CacheRedis.DB,
		KeyPrefix: c.IpGeo.CacheRedis.KeyPrefix,
	})
	if err != nil {
		_ = client.Close()
		logx.Must(err)
	}

	deps := service.NewDeps(client, cch)
	ctx := context.Background()
	if err := deps.IpGeo.SeedProvidersIfEmpty(ctx, toSeedProviders(c)); err != nil {
		closeCache(cch)
		_ = client.Close()
		logx.Must(err)
	}
	if err := deps.IpGeo.FillEmptyProviderTags(ctx); err != nil {
		closeCache(cch)
		_ = client.Close()
		logx.Must(err)
	}

	ttl := time.Duration(0)
	if c.IpGeo.CacheTTL > 0 {
		ttl = time.Duration(c.IpGeo.CacheTTL) * time.Hour
	}
	lookup, err := ipgeo.New(ipgeo.Conf{
		Client:      client,
		Cache:       cch,
		CacheTTL:    ttl,
		ListEnabled: deps.IpGeo.ListEnabledProviders,
	})
	if err != nil {
		closeCache(cch)
		_ = client.Close()
		logx.Must(err)
	}

	return &ServiceContext{
		Config: c,
		Deps:   deps,
		IpGeo:  lookup,
	}
}

// Close 关闭 Lookup、缓存与 Ent Client。
func (s *ServiceContext) Close() error {
	if s == nil {
		return nil
	}
	var first error
	if s.IpGeo != nil {
		if err := s.IpGeo.Close(); err != nil {
			first = err
		}
	}
	if s.Deps != nil {
		if s.Deps.IpGeo != nil {
			if err := closeCache(s.Deps.IpGeo.Cache); err != nil && first == nil {
				first = err
			}
		}
		if s.Deps.Client != nil {
			if err := s.Deps.Client.Close(); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}

func closeCache(cch any) error {
	if c, ok := cch.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

// toSeedProviders 把 yaml IpGeo.Providers 转成领域 Provider（仅启动种子用）。
func toSeedProviders(c config.Config) []geo.Provider {
	out := make([]geo.Provider, 0, len(c.IpGeo.Providers))
	for _, p := range c.IpGeo.Providers {
		out = append(out, geo.Provider{
			Name:       p.Name,
			Driver:     p.Driver,
			Tag:        p.Tag,
			IsEnabled:  p.IsEnabled,
			Priority:   p.Priority,
			Endpoint:   p.Endpoint,
			AuthType:   p.AuthType,
			AccessKey:  p.AccessKey,
			SecretKey:  p.SecretKey,
			TimeoutSec: p.TimeoutSec,
			FieldMap:   p.FieldMap,
			Headers:    p.Headers,
		})
	}
	return out
}
