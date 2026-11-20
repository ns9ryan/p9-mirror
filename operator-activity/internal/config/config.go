package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-activity/pkg/database"
	taskconfig "oa.98ent.com/p9/operator-activity/task/config"
)

type Config struct {
	zrpc.RpcServerConf
	// 数据库配置
	DatabaseConf database.DatabaseConf
	// 后台CRM RPC配置
	CrmRpc zrpc.RpcClientConf
	// 后台基础服务 RPC配置
	OperatorBaseRpc zrpc.RpcClientConf
	// Kafka配置
	KqBrokers []string
	// Redis配置
	RedisAddr string
	// Temporal配置
	TemporalInfo TemporalInfo
	// 是否启动活动引擎
	EnableActivityEngine bool

	// 是否启用任务引擎
	EnableTaskEngine bool
	TaskConfig       TaskConfig
}

// TemporalInfo 配置
type TemporalInfo struct {
	Address   string `json:"address"`
	Namespace string `json:"namespace"`
}

type TaskConfig struct {
	KqBrokers       []string
	RedisConf       taskconfig.RedisConf
	EnableTaskTimer bool
}

// IsDebug 判断是否为开发或测试模式
func (c Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}
