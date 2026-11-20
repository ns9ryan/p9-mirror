package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-activity/pkg/database"
)

type Config struct {
	zrpc.RpcServerConf
	// 服务名称
	Name string
	// 数据库配置
	DatabaseConf database.DatabaseConf
	KqBrokers    []string
	RedisConf    RedisConf
	// 游戏服务 RPC 配置
	OperatorGameRpc zrpc.RpcClientConf
	OperatorBaseRpc zrpc.RpcClientConf

	// 是否开启定时器
	EnableTaskTimer bool
}

func (c Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}

type RedisConf struct {
	Address  string
	Password string
	DB       int
}
