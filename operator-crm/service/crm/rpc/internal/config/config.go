package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-crm/pkg/database"
)

type Config struct {
	OperatorCoreRpc zrpc.RpcClientConf `json:",optional"`
	IntegrationRpc  zrpc.RpcClientConf `json:",optional"`
	PlatformBaseRpc zrpc.RpcClientConf `json:",optional"`
	FieldCrypto     struct {
		EncryptionKey string
		LookupKey     string
	} `json:",optional"`

	zrpc.RpcServerConf
	// 数据库配置
	DatabaseConf database.DatabaseConf
	// VIP素材公开前缀，包含公开桶路径；只在请求时校验，不使用内部上传地址。
	ResourceBaseURL string `json:",optional"`
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
	// Kafka brokers 配置
	Kafka struct {
		Brokers []string
	} `json:",optional"`
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
