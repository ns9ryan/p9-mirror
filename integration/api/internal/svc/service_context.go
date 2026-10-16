package svc

import (
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/common/coreadapt"
	"oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/integration/api/internal/config"
	"oa.98ent.com/p9/integration/api/internal/locales"
	"oa.98ent.com/p9/integration/rpc/integrationclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext HTTP API 依赖注入。
// 本层不直连业务库：鉴权走 core RPC，业务走 integration RPC。
type ServiceContext struct {
	Config      config.Config                 // API 配置
	Core        coreclient.Core               // core gRPC（catalog / CheckToken / Enforce）
	Integration integrationclient.Integration // integration gRPC 客户端
	Jwt         rest.Middleware               // CheckToken
	Authority   rest.Middleware               // Casbin Enforce
	ActionLog   rest.Middleware               // 操作日志中间件
	ErrorLog    rest.Middleware               // 错误日志中间件
	Trans       *i18n.Translator              // 多语言翻译器
}

// NewServiceContext 构造 API 依赖：core/integration 客户端与鉴权、日志中间件。
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建API翻译器
	trans, err := i18n.New(c.I18n, locales.FS)
	logx.Must(err)

	cli := zrpc.MustNewClient(c.IntegrationRpc)
	coreCli := coreclient.NewCore(zrpc.MustNewClient(c.CoreRpc))
	auth := coreadapt.Auth(coreCli)
	coreadapt.SetDictLoader(coreCli)

	return &ServiceContext{
		Config:      c,
		Core:        coreCli,
		Trans:       trans,
		Integration: integrationclient.NewIntegration(cli),
		Jwt:         middleware.JWT(auth),
		Authority:   middleware.Authority(auth),
		ActionLog:   middleware.ActionLog(coreadapt.ActionRecorder(coreCli)),
		ErrorLog:    middleware.ErrorLog(c.Name, coreadapt.ErrorRecorder(coreCli)),
	}
}
