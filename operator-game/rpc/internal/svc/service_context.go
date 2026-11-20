package svc

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-game/rpc/ent"
	"oa.98ent.com/p9/operator-game/rpc/ent/migrate"
	"oa.98ent.com/p9/operator-game/rpc/internal/cache"
	"oa.98ent.com/p9/operator-game/rpc/internal/config"
	"oa.98ent.com/p9/operator-game/rpc/internal/dao"
	"oa.98ent.com/p9/platform-game/pkg/grpc_client"
)

type ServiceContext struct {
	Config                 config.Config
	DB                     *ent.Client    // Ent数据库客户端
	DAOManager             *dao.Manager   // DAO管理器
	CacheManager           *cache.Manager // 缓存管理器
	PlatformGameGrpcClient *grpc_client.GameClientManager
	Core                   coreclient.Core // Core RPC客户端
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建数据库驱动
	driver, err := c.DatabaseConf.NewDriver()
	logx.Must(err)

	// 创建Ent客户端配置
	entOpts := []ent.Option{
		ent.Log(logx.Info), // 使用go-zero日志输出SQL
		ent.Driver(driver), // 设置数据库驱动
	}

	// 开发和测试环境开启Ent调试模式
	if c.Mode == service.DevMode || c.Mode == service.TestMode {
		entOpts = append(entOpts, ent.Debug())
	}

	// 创建Ent数据库客户端
	db := ent.NewClient(entOpts...)

	// 创建平台游戏服务RPC客户端
	gameGrpcClient, err := grpc_client.NewGameClientManager(c.PlatformGameRpcConf)

	// 创建 Core 客户端并初始化鉴权
	coreClient := zrpc.MustNewClient(c.CoreRpc)
	coreCli := coreclient.NewCore(coreClient)
	return &ServiceContext{
		Config:                 c,
		DB:                     db,
		DAOManager:             dao.NewManager(db),
		CacheManager:           cache.NewManager(),
		PlatformGameGrpcClient: gameGrpcClient,
		Core:                   coreCli,
	}
}

func (s *ServiceContext) MustMigrate() {
	ctx := context.Background()

	// 根据 Ent Schema 自动创建或更新数据库结构
	logx.Must(
		s.DB.Schema.Create(
			ctx,
			migrate.WithForeignKeys(false), // 不创建数据库外键
			migrate.WithDropIndex(true),    // 允许删除废弃索引
		),
	)

}
