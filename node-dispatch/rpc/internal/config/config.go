package config

import (
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/node-dispatch/pkg/database"
)

// CallbackConf HTTP回调配置
type CallbackConf struct {
	TaskResultURL string // 任务结果回调地址
	Secret        string // 回调认证密钥
	Timeout       int64  // HTTP请求超时时间, 毫秒
}

type Config struct {
	zrpc.RpcServerConf

	// WebSocket监听地址
	WebSocketListenOn string

	// HTTP回调配置
	Callback CallbackConf

	// 数据库配置
	DatabaseConf database.DatabaseConf
}
