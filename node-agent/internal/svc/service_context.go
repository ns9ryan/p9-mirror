package svc

import (
	"context"

	coreservice "oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/node-agent/internal/config"
	"oa.98ent.com/p9/node-agent/internal/domain"
	"oa.98ent.com/p9/node-agent/internal/task"
	"oa.98ent.com/p9/node-agent/internal/websocket"
	operatorbaseservice "oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	crminitservice "oa.98ent.com/p9/operator-crm/service/crm/rpc/client/init"
	operatorgameservice "oa.98ent.com/p9/operator-game/rpc/client/publishdataservice"
	platformoperatorservice "oa.98ent.com/p9/platform-operator/rpc/client/operatorservice"

	"github.com/zeromicro/go-zero/core/logx"
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
	// 仅创建可选CRM客户端；实际初始化在发布分站任务中执行，不在节点启动时写库。
	taskService.CrmInitRpc = newCrmInitClient(c.OperatorCrmRpc)

	// 初始化 Nginx 管理器
	ctx := context.Background()
	domainManager := domain.NewManager(ctx, c)
	taskService.DomainManager = domainManager

	// 创建节点WebSocket客户端
	webSocketClient := websocket.NewClient(c, taskService)

	return &ServiceContext{
		Config:    c,               // 服务配置
		Task:      taskService,     // 调度任务执行服务
		WebSocket: webSocketClient, // 节点WebSocket客户端
	}
}

// newCrmInitClient 参考CRM可选客户端的写法，缺配置或创建失败不增加节点启动必需依赖。
func newCrmInitClient(conf *zrpc.RpcClientConf) crminitservice.Init {
	if conf == nil || len(conf.Endpoints) == 0 {
		logx.Error("当前节点CRM初始化RPC缺少Endpoints直连配置，发布分站时CRM初始化不可用")
		return nil
	}
	// 强制非阻塞直连，不等待CRM上线，也不使用可能同步等待的etcd发现。
	clientConf := *conf
	clientConf.NonBlock = true
	client, err := zrpc.NewClient(clientConf)
	if err != nil {
		logx.Errorw("创建当前节点CRM初始化RPC客户端失败", logx.Field("error", err))
		return nil
	}
	return crminitservice.NewInit(client)
}
