// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-base/api/internal/config"
	"oa.98ent.com/p9/operator-base/rpc/client/pingservice"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	// 服务配置
	Config config.Config // 服务配置

	// Operator Base RPC
	PingRpc pingservice.PingService // Ping RPC
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// ============================== Operator Base RPC ==============================

	// 创建Operator Base RPC客户端
	operatorBaseClient := zrpc.MustNewClient(c.OperatorBaseRpc)

	// ============================== Service Context ==============================

	return &ServiceContext{
		// 服务配置
		Config: c, // 服务配置

		// Operator Base RPC
		PingRpc: pingservice.NewPingService(operatorBaseClient), // Ping RPC
	}
}
