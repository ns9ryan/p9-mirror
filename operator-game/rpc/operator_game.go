package main

import (
	"flag"
	"fmt"

	"oa.98ent.com/p9/operator-game/rpc/internal/config"
	gamecategoryserviceServer "oa.98ent.com/p9/operator-game/rpc/internal/server/gamecategoryservice"
	gamechannelserviceServer "oa.98ent.com/p9/operator-game/rpc/internal/server/gamechannelservice"
	gameproviderserviceServer "oa.98ent.com/p9/operator-game/rpc/internal/server/gameproviderservice"
	gameserviceServer "oa.98ent.com/p9/operator-game/rpc/internal/server/gameservice"
	publishdataservice "oa.98ent.com/p9/operator-game/rpc/internal/server/publishdataservice"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/operator_game.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		operator_game.RegisterGameServiceServer(grpcServer, gameserviceServer.NewGameServiceServer(ctx))
		operator_game.RegisterGameCategoryServiceServer(grpcServer, gamecategoryserviceServer.NewGameCategoryServiceServer(ctx))
		operator_game.RegisterGameProviderServiceServer(grpcServer, gameproviderserviceServer.NewGameProviderServiceServer(ctx))
		operator_game.RegisterGameChannelServiceServer(grpcServer, gamechannelserviceServer.NewGameChannelServiceServer(ctx))
		operator_game.RegisterPublishDataServiceServer(grpcServer, publishdataservice.NewPublishDataServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
