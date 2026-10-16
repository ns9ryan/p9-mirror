// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/agent-admin-api/internal/config"
	"oa.98ent.com/p9/agent-admin-api/internal/locales"
	"oa.98ent.com/p9/agent-admin-api/internal/middleware"
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/common/coreadapt"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/authservice"
)

type ServiceContext struct {
	Config          config.Config
	Trans           *i18n.Translator
	OperatorCoreRpc coreclient.Core
	AgentAuthRpc    authservice.AuthService
	Jwt             rest.Middleware // CRM JWT，仅菜单等受保护路由使用
	JwtWithoutError rest.Middleware // CRM JWTWithoutError，不返回错误
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建API翻译器
	trans, err := i18n.New(c.I18n, locales.FS)
	logx.Must(err)

	// 创建分站Core RPC客户端
	operatorCoreCli := zrpc.MustNewClient(c.OperatorCoreRpc)
	operatorCoreRpc := coreclient.NewCore(operatorCoreCli)
	// 设置多语言字典加载器
	coreadapt.SetDictLoader(operatorCoreRpc)

	// 创建后台CRM RPC客户端
	authCli := zrpc.MustNewClient(c.CrmRpc)
	authRpc := authservice.NewAuthService(authCli)

	jwtCli := middleware.JwtClient(authRpc)

	return &ServiceContext{
		Config:          c,
		Trans:           trans,
		OperatorCoreRpc: operatorCoreRpc,
		AgentAuthRpc:    authRpc,
		Jwt:             middleware.NewJwtMiddleware(jwtCli),
		JwtWithoutError: middleware.NewJwtWithoutErrorMiddleware(jwtCli),
	}
}
