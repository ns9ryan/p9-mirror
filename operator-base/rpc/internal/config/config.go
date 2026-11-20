package config

import (
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-base/pkg/database"
)

// Config RPC服务配置
type Config struct {
	zrpc.RpcServerConf

	// 数据库配置
	DatabaseConf database.DatabaseConf
	// 新分站的默认 Logo，支持公开 URL 或 defaults/ 下的 Key；未配置时为空。
	DefaultSiteLogoURL string `json:",optional"`

	// Logo 公开访问前缀，包含必要的桶名或路径；为空时兼容原有完整地址存储。
	SiteLogoBaseURL string `json:",optional"`
	// 更换访问地址后仍需识别的旧前缀，仅用于将已知旧地址转换成对象 Key。
	SiteLogoLegacyBaseURLs []string `json:",optional"`

	// 详情查询依赖，未配置时不影响原有初始化和编码查询业务。
	OperatorCoreRpc *zrpc.RpcClientConf `json:",optional"` // 当前节点 Core，查询本站语言
	PlatformBaseRpc *zrpc.RpcClientConf `json:",optional"` // 总网基础资料
}
