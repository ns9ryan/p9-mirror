package main

import (
	"flag"
	"fmt"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/config"
	memberWalletServer "oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/server/memberwallet"
	pingServer "oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/server/ping"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/member_wallet.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.DB.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 健康检查服务
		memberwalletrpc.RegisterPingServer(grpcServer, pingServer.NewPingServer(ctx))
		// 会员钱包资金服务
		memberwalletrpc.RegisterMemberWalletServer(grpcServer, memberWalletServer.NewMemberWalletServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
