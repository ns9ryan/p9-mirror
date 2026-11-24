// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/common/i18n"
)

// MemberExportConfig is optional; missing dependencies disable only export.
type MemberExportConfig struct {
	RedisAddr     string `json:",optional"`
	RedisPassword string `json:",optional"`
	RedisDB       int    `json:",optional"`
}
type Config struct {
	MemberExport MemberExportConfig `json:",optional"`

	rest.RestConf
	// 分站Core RPC配置
	OperatorCoreRpc zrpc.RpcClientConf
	// 后台基础服务 RPC配置
	OperatorBaseRpc zrpc.RpcClientConf
	// 后台CRM RPC配置
	CrmRpc zrpc.RpcClientConf
	// 会员钱包RPC配置
	MemberWalletRpc zrpc.RpcClientConf
	// 活动服务 RPC配置
	ActivityRpc zrpc.RpcClientConf
	// 总站基础RPC配置；VIP默认图标仅在查询请求中调用。
	PlatformBaseRpc zrpc.RpcClientConf `json:",optional"`
	// 第三方服务集成 RPC配置
	IntegrationRpc zrpc.RpcClientConf
	// 分站 OSS API 根地址；仅上传时使用，未配置不影响原有服务启动。
	OssApiBaseURL string `json:",optional"`
	// 国际化配置
	I18n i18n.Config

	Kafka struct {
		Brokers []string `json:",optional"`
	} `json:",optional"`
}

// IsDebug 是否为调试模式
func (c *Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}

// IsDev 是否为开发模式
func (c *Config) IsDev() bool {
	return c.Mode == service.DevMode
}
