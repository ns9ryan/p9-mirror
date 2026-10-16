package main

import (
	"flag"
	"fmt"

	_ "oa.98ent.com/p9/operator-crm/ent/runtime"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/config"
	agentserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/agentservice"
	authserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/authservice"
	pingserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/pingservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/crm.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		crmrpc.RegisterPingServiceServer(grpcServer, pingserviceServer.NewPingServiceServer(ctx))
		crmrpc.RegisterAuthServiceServer(grpcServer, authserviceServer.NewAuthServiceServer(ctx))
		crmrpc.RegisterAgentServiceServer(grpcServer, agentserviceServer.NewAgentServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
