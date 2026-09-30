package grpc_client

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/operator-game/rpc/client/gamecategoryservice"
	"oa.98ent.com/p9/operator-game/rpc/client/gamechannelservice"
	"oa.98ent.com/p9/operator-game/rpc/client/gameproviderservice"
	"oa.98ent.com/p9/operator-game/rpc/client/gameservice"
	"oa.98ent.com/p9/operator-game/rpc/client/publishdataservice"
)

// ClientManager gRPC客户端管理器
type GameClientManager struct {
	gameServiceClient         gameservice.GameService
	gameCategoryServiceClient gamecategoryservice.GameCategoryService
	gameProviderServiceClient gameproviderservice.GameProviderService
	gameChannelServiceClient  gamechannelservice.GameChannelService
	publishDataServiceClient  publishdataservice.PublishDataService
}

// 创建新的游戏gRPC客户端管理器
func NewOperatorGameClientManager(cfg zrpc.RpcClientConf) (*GameClientManager, error) {
	logx.Infof("create version 1.0.0 game client manager")
	client := zrpc.MustNewClient(
		cfg,
		zrpc.WithUnaryClientInterceptor(UnaryClientInterceptor),
	)

	return &GameClientManager{
		gameServiceClient:         gameservice.NewGameService(client),
		gameCategoryServiceClient: gamecategoryservice.NewGameCategoryService(client),
		gameProviderServiceClient: gameproviderservice.NewGameProviderService(client),
		gameChannelServiceClient:  gamechannelservice.NewGameChannelService(client),
		publishDataServiceClient:  publishdataservice.NewPublishDataService(client),
	}, nil
}

// GetGameServiceClient 获取游戏服务客户端
func (cm *GameClientManager) GetGameServiceClient() gameservice.GameService {
	return cm.gameServiceClient
}

// GetGameCategoryServiceClient 获取游戏分类服务客户端
func (cm *GameClientManager) GetGameCategoryServiceClient() gamecategoryservice.GameCategoryService {
	return cm.gameCategoryServiceClient
}

// GetGameProviderServiceClient 获取游戏供应商服务客户端
func (cm *GameClientManager) GetGameProviderServiceClient() gameproviderservice.GameProviderService {
	return cm.gameProviderServiceClient
}

// GetGameChannelServiceClient 获取游戏渠道服务客户端
func (cm *GameClientManager) GetGameChannelServiceClient() gamechannelservice.GameChannelService {
	return cm.gameChannelServiceClient
}

// GetPublishDataServiceClient 获取发布数据服务客户端
func (cm *GameClientManager) GetPublishDataServiceClient() publishdataservice.PublishDataService {
	return cm.publishDataServiceClient
}
