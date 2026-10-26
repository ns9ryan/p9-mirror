package task

import (
	coreservice "oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/node-agent/internal/domain"
	operatorbaseservice "oa.98ent.com/p9/operator-base/rpc/client/operatorservice"
	crminitservice "oa.98ent.com/p9/operator-crm/service/crm/rpc/client/init"
	operatorgameservice "oa.98ent.com/p9/operator-game/rpc/client/publishdataservice"
	platformoperatorservice "oa.98ent.com/p9/platform-operator/rpc/client/operatorservice"
)

// Service 调度任务执行服务
type Service struct {
	platformOperatorRpc platformoperatorservice.OperatorService // 总网分站RPC
	operatorBaseRpc     operatorbaseservice.OperatorService     // 当前节点分站基础RPC
	platformCoreRpc     coreservice.Core                        // 总网Core RPC
	operatorCoreRpc     coreservice.Core                        // 分站Core RPC
	operatorGameRpc     operatorgameservice.PublishDataService  // 当前节点游戏RPC
	operatorCrmRpc      crminitservice.Init                     // 当前节点CRM初始化RPC
	DomainManager       *domain.Manager                         // Nginx 管理器
}

// NewService 创建调度任务执行服务
func NewService(
	platformOperatorRpc platformoperatorservice.OperatorService,
	operatorBaseRpc operatorbaseservice.OperatorService,
	platformCoreRpc coreservice.Core,
	operatorCoreRpc coreservice.Core,
	operatorGameRpc operatorgameservice.PublishDataService,
	operatorCrmRpc crminitservice.Init,
) *Service {
	return &Service{
		platformOperatorRpc: platformOperatorRpc,
		operatorBaseRpc:     operatorBaseRpc,
		platformCoreRpc:     platformCoreRpc,
		operatorCoreRpc:     operatorCoreRpc,
		operatorGameRpc:     operatorGameRpc,
		operatorCrmRpc:      operatorCrmRpc,
	}
}
