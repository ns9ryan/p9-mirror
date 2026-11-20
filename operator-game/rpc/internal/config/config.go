package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 应用配置结构
type Config struct {
	zrpc.RpcServerConf

	DatabaseConf DatabaseConf
	// 总网游戏服务 gRPC 配置（用于同步已发布数据）
	PlatformGameRpcConf zrpc.RpcClientConf `json:"platformGameRpcConf,optional" yaml:"PlatformGameRpcConf"`

	CoreRpc zrpc.RpcClientConf `json:"coreRpc,optional" yaml:"CoreRpc"`
}
