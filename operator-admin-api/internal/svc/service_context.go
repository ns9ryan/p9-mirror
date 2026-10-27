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
	"oa.98ent.com/p9/integration/rpc/integrationclient"
	"oa.98ent.com/p9/operator-admin-api/internal/config"
	"oa.98ent.com/p9/operator-admin-api/internal/locales"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/agentservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/memberautolevel"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/memberfixedlevel"
)

// ServiceContext 分站后台服务上下文
type ServiceContext struct {
	// 基础依赖
	Config config.Config    // 服务配置
	Trans  *i18n.Translator // 多语言翻译器

	// 基础服务RPC
	OperatorCoreRpc coreclient.Core                 // Core RPC
	OperatorBaseRpc operatorservice.OperatorService // 分站基础RPC

	// CRM业务RPC
	AgentRpc            agentservice.AgentService         // 分层代理RPC
	MemberAutoLevelRpc  memberautolevel.MemberAutoLevel   // 会员自动层级RPC
	MemberFixedLevelRpc memberfixedlevel.MemberFixedLevel // 会员固定层级RPC

	// 第三方服务集成 RPC
	IntegrationRpc integrationclient.Integration

	// 认证权限
	Jwt       rest.Middleware // JWT认证
	Authority rest.Middleware // 权限校验

	// 日志
	ActionLog rest.Middleware // 操作日志
	ErrorLog  rest.Middleware // 错误日志
}

// NewServiceContext 创建分站后台服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化API翻译器
	trans, err := i18n.New(c.I18n, locales.FS)
	logx.Must(err)

	// 创建Core RPC客户端
	coreClient := zrpc.MustNewClient(c.OperatorCoreRpc)
	coreRpc := coreclient.NewCore(coreClient)

	// 设置Core字典加载器和认证适配器
	coreadapt.SetDictLoader(coreRpc)
	auth := coreadapt.Auth(coreRpc)

	// 创建业务RPC客户端
	baseClient := zrpc.MustNewClient(c.OperatorBaseRpc)
	crmClient := zrpc.MustNewClient(c.CrmRpc)

	// 创建第三方服务集成客户端
	integrationClient := zrpc.MustNewClient(c.IntegrationRpc)

	return &ServiceContext{
		Config: c,
		Trans:  trans,

		OperatorCoreRpc: coreRpc,
		OperatorBaseRpc: operatorservice.NewOperatorService(baseClient),

		AgentRpc:            agentservice.NewAgentService(crmClient),
		MemberAutoLevelRpc:  memberautolevel.NewMemberAutoLevel(crmClient),
		MemberFixedLevelRpc: memberfixedlevel.NewMemberFixedLevel(crmClient),

		IntegrationRpc: integrationclient.NewIntegration(integrationClient),

		Jwt:       middleware.JWT(auth),
		Authority: middleware.Authority(auth),

		ActionLog: middleware.ActionLog(coreadapt.ActionRecorder(coreRpc)),
		ErrorLog:  middleware.ErrorLog(c.Name, coreadapt.ErrorRecorder(coreRpc)),
	}
}
