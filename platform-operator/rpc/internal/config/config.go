package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/platform-operator/pkg/database"
)

type Config struct {
	zrpc.RpcServerConf

	// Node Dispatch RPC配置
	NodeDispatchRpc zrpc.RpcClientConf

	// 数据库配置
	DatabaseConf database.DatabaseConf
}
