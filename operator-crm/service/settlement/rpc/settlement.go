package main

import (
	"flag"
	"fmt"

	"oa.98ent.com/p9/operator-crm/service/settlement/rpc/internal/config"
	pingserviceServer "oa.98ent.com/p9/operator-crm/service/settlement/rpc/internal/server/pingservice"
	"oa.98ent.com/p9/operator-crm/service/settlement/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/settlement/rpc/pb/settlementrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/settlement.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		settlementrpc.RegisterPingServiceServer(grpcServer, pingserviceServer.NewPingServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
