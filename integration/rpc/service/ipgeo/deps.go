// Package ipgeo IP/GEO 域业务层（LocalIP / 策略 / Provider）。
// 查询编排在 internal/ipgeo/lookup；本包只做管理端 CRUD 与选源列表。
package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/ent"
)

// Cache 写本地 IP 后按 IP 失效查询缓存；由 svc 注入（table/memory/redis 实现）。
type Cache interface {
	DeleteByIP(ctx context.Context, ip string) error
}

// Service IP/GEO 域依赖与用例入口。
type Service struct {
	Client *ent.Client // 业务库（可与其它域共用同一 Client）
	Cache  Cache       // 可空
}

// New 构造 IP/GEO 域服务。
func New(client *ent.Client, cch Cache) *Service {
	return &Service{Client: client, Cache: cch}
}
