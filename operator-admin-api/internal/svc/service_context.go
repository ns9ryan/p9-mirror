// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/core/common/coreadapt"
	"oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-admin-api/internal/config"
	"oa.98ent.com/p9/operator-admin-api/internal/locales"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/agentservice"
)

type ServiceContext struct {
	Config          config.Config
	Trans           *i18n.Translator
	OperatorCoreRpc coreclient.Core
	OperatorBaseRpc operatorservice.OperatorService
	AgentRpc        agentservice.AgentService
	Jwt             rest.Middleware // Core JWT
	Authority       rest.Middleware // Casbin 权限
	ActionLog       rest.Middleware // 操作日志
	ErrorLog        rest.Middleware // 错误日志
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建API翻译器
	trans, err := i18n.New(c.I18n, locales.FS)
	logx.Must(err)

	// 创建分站Core RPC客户端
	operatorCoreCli := zrpc.MustNewClient(c.OperatorCoreRpc)
	operatorCoreRpc := coreclient.NewCore(operatorCoreCli)

	// 创建后台基础服务 RPC客户端
	operatorBaseCli := zrpc.MustNewClient(c.OperatorBaseRpc)
	operatorBaseRpc := operatorservice.NewOperatorService(operatorBaseCli)

	// 设置多语言字典加载器
	coreadapt.SetDictLoader(operatorCoreRpc)

	// 权限中间件适配器
	authAdapter := coreadapt.Auth(operatorCoreRpc)

	// 创建后台 CRM Agent RPC 客户端
	crmCli := zrpc.MustNewClient(c.CrmRpc)
	agentRpc := agentservice.NewAgentService(crmCli)

	return &ServiceContext{
		Config:          c,
		Trans:           trans,
		OperatorCoreRpc: operatorCoreRpc,
		OperatorBaseRpc: operatorBaseRpc,
		AgentRpc:        agentRpc,
		Jwt:             middleware.JWT(authAdapter),
		Authority:       middleware.Authority(authAdapter),
		ActionLog:       middleware.ActionLog(coreadapt.ActionRecorder(operatorCoreRpc)),
		ErrorLog:        middleware.ErrorLog(c.Name, coreadapt.ErrorRecorder(operatorCoreRpc)),
	}
}
