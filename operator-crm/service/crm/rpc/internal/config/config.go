package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/operator-crm/pkg/database"
)

type Config struct {
	zrpc.RpcServerConf
	// 数据库配置
	DatabaseConf database.DatabaseConf
	// 代理后台 JWT配置
	AgentAdminJwt struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshSecret string
		RefreshExpire int64
	}
	// 令牌黑名单 Redis 配置
	DataRedis struct {
		Host string
		Pass string
		DB   int
	}
}

func (c Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}

func (c Config) IsDev() bool {
	return c.Mode == service.DevMode
}

func (c Config) IsTest() bool {
	return c.Mode == service.TestMode
}
