package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 节点代理配置
type Config struct {
	NodeCode   string       // 节点业务编码
	AuthSecret string       // 节点认证密钥
	Dispatch   DispatchConf // 调度中心连接配置
	Nginx      NginxConf    // Nginx 配置

	PlatformOperatorRpc zrpc.RpcClientConf // 总网分站RPC配置
	OperatorBaseRpc     zrpc.RpcClientConf // 当前节点分站基础RPC配置
	PlatformCoreRpc     zrpc.RpcClientConf // 总网Core RPC配置
	OperatorCoreRpc     zrpc.RpcClientConf // 分站Core RPC配置
	OperatorGameRpc     zrpc.RpcClientConf // 当前节点游戏RPC配置
	OperatorCrmRpc      zrpc.RpcClientConf // 当前节点CRM RPC配置
}

// DispatchConf 调度中心连接配置
type DispatchConf struct {
	WebSocketURL string // WebSocket连接地址
}

// NginxConf Nginx 配置
type NginxConf struct {
	ConfDirPath   string // Nginx 配置目录路径（例如：/home/bbs/docker-compose/deploy/nginx/conf.d）
	ContainerName string // Nginx 容器名称（默认：p9-nginx）
	PortMap       NginxPortMap
}

// NginxPortMap Nginx 端口映射
type NginxPortMap struct {
	Core     int // Core API 端口
	Base     int // Base API 端口
	Game     int // Game API 端口
	Operator int // Operator API 端口
	Dispatch int // Dispatch API 端口
}
