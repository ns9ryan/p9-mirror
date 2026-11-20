package config

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/platform-base/pkg/database"
)

type Config struct {
	// ListCacheTTL controls the TTL of nonempty resource and nickname lists.
	ListCacheTTL time.Duration `json:",default=24h"`

	DataRedis struct {
		Host string
		Pass string
		DB   int
	} `json:",optional"`

	zrpc.RpcServerConf

	// 数据库配置
	DatabaseConf database.DatabaseConf

	// VIP、头像和头像框共用的公开访问前缀，包含桶路径；为空时仅兼容已有完整 URL。
	// 仅在处理素材请求时使用，不增加服务启动依赖或读取其他服务配置。
	ResourceBaseURL string `json:",optional"`
}
