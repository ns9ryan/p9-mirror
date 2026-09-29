// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/common/i18n"
)

type Config struct {
	rest.RestConf

	// Node Dispatch RPC配置
	NodeDispatchRpc zrpc.RpcClientConf

	// Core RPC配置
	CoreRpc zrpc.RpcClientConf

	// 国际化配置
	I18n i18n.Config
}

// IsDebug 是否为调试模式
func (c *Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}

// IsDev 是否为开发模式
func (c *Config) IsDev() bool {
	return c.Mode == service.DevMode
}
