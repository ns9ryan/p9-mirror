// Package service 业务层根包（对齐 core/rpc/service）。
//
// 按域分子包：service/ipgeo、以后 service/message 等。
// 本文件只放跨域共享的 Deps 聚合；各域用例在子包内。
package service

import (
	"oa.98ent.com/p9/integration/rpc/ent"
	ipgeosvc "oa.98ent.com/p9/integration/rpc/service/ipgeo"
)

// Deps 跨域业务依赖聚合。
// 当前共用一套 PostgreSQL Client；分库时可为各域注入不同 Client。
type Deps struct {
	Client *ent.Client       // 当前共用业务库
	IpGeo  *ipgeosvc.Service // IP/GEO 域
	// Message *messagesvc.Service // 示例：后续消息域
}

// NewDeps 组装各域服务；ipGeoCache 可空（不写缓存失效）。
func NewDeps(client *ent.Client, ipGeoCache ipgeosvc.Cache) *Deps {
	return &Deps{
		Client: client,
		IpGeo:  ipgeosvc.New(client, ipGeoCache),
	}
}
