// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/core/common/coreadapt"
	coremiddleware "oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-base/api/internal/config"
	"oa.98ent.com/p9/operator-base/api/internal/middleware"
	"oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	"oa.98ent.com/p9/operator-base/rpc/client/pingservice"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	// 服务配置
	Config config.Config // 服务配置
	Core   coreclient.Core

	// Operator Base RPC
	PingRpc     pingservice.PingService         // Ping RPC
	OperatorRpc operatorservice.OperatorService // 站点详情 RPC
	SiteAuth    rest.Middleware                 // 登录、站点身份及接口权限校验
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// ============================== Operator Base RPC ==============================

	// 创建Operator Base RPC客户端
	operatorBaseClient := zrpc.MustNewClient(c.OperatorBaseRpc)
	var coreCli coreclient.Core
	var auth coremiddleware.Client
	if c.CoreRpc != nil {
		coreConfig := *c.CoreRpc
		coreConfig.NonBlock = true
		coreCli = coreclient.NewCore(zrpc.MustNewClient(coreConfig))
		auth = coreadapt.Auth(coreCli)
	}

	// ============================== Service Context ==============================

	return &ServiceContext{
		// 服务配置
		Config: c, // 服务配置
		Core:   coreCli,

		// Operator Base RPC
		PingRpc:     pingservice.NewPingService(operatorBaseClient), // Ping RPC
		OperatorRpc: operatorservice.NewOperatorService(operatorBaseClient),
		SiteAuth:    middleware.NewSiteAuth(auth),
	}
}
