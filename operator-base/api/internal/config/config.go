// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// Operator Base RPC配置
	OperatorBaseRpc zrpc.RpcClientConf

	// 本站 Core，负责目录登记和请求鉴权；按当前启动约定，缺失或登记失败时 API 不启动。
	CoreRpc *zrpc.RpcClientConf `json:",optional"`
}
