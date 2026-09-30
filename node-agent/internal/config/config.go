package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 节点代理配置
type Config struct {
	NodeCode   string       // 节点业务编码
	AuthSecret string       // 节点认证密钥
	Dispatch   DispatchConf // 调度中心连接配置

	PlatformOperatorRpc zrpc.RpcClientConf // 总网分站RPC配置
	OperatorBaseRpc     zrpc.RpcClientConf // 当前节点分站基础RPC配置
	PlatformCoreRpc     zrpc.RpcClientConf // 总网Core RPC配置
	OperatorCoreRpc     zrpc.RpcClientConf // 分站Core RPC配置
	OperatorGameRpc     zrpc.RpcClientConf // 当前节点游戏RPC配置
}

// DispatchConf 调度中心连接配置
type DispatchConf struct {
	WebSocketURL string // WebSocket连接地址
}
