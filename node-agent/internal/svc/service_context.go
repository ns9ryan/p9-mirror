package svc

import (
	coreservice "oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/node-agent/internal/config"
	"oa.98ent.com/p9/node-agent/internal/task"
	"oa.98ent.com/p9/node-agent/internal/websocket"
	operatorbaseservice "oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	operatorgameservice "oa.98ent.com/p9/operator-game/rpc/client/publishdataservice"
	platformoperatorservice "oa.98ent.com/p9/platform-operator/rpc/client/operatorservice"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config    config.Config     // 服务配置
	Task      *task.Service     // 调度任务执行服务
	WebSocket *websocket.Client // 节点WebSocket客户端
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建总网分站RPC客户端
	platformOperatorClient := zrpc.MustNewClient(c.PlatformOperatorRpc)
	platformOperatorRpc := platformoperatorservice.NewOperatorService(platformOperatorClient)

	// 创建当前节点分站基础RPC客户端
	operatorBaseClient := zrpc.MustNewClient(c.OperatorBaseRpc)
	operatorBaseRpc := operatorbaseservice.NewOperatorService(operatorBaseClient)

	// 创建总网Core RPC客户端
	platformCoreClient := zrpc.MustNewClient(c.PlatformCoreRpc)
	platformCoreRpc := coreservice.NewCore(platformCoreClient)

	// 创建分站Core RPC客户端
	operatorCoreClient := zrpc.MustNewClient(c.OperatorCoreRpc)
	operatorCoreRpc := coreservice.NewCore(operatorCoreClient)

	// 创建当前节点游戏RPC客户端
	operatorGameClient := zrpc.MustNewClient(c.OperatorGameRpc)
	operatorGameRpc := operatorgameservice.NewPublishDataService(operatorGameClient)

	// 创建调度任务执行服务
	taskService := task.NewService(platformOperatorRpc, operatorBaseRpc, platformCoreRpc, operatorCoreRpc, operatorGameRpc)

	// 创建节点WebSocket客户端
	webSocketClient := websocket.NewClient(c, taskService)

	return &ServiceContext{
		Config:    c,               // 服务配置
		Task:      taskService,     // 调度任务执行服务
		WebSocket: webSocketClient, // 节点WebSocket客户端
	}
}
