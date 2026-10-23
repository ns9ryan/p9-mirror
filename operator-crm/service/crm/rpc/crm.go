package main

import (
	"flag"
	"fmt"

	_ "oa.98ent.com/p9/operator-crm/ent/runtime"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/config"
	agentserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/agentservice"
	authserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/authservice"
	initServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/init"
	memberautolevelServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/memberautolevel"
	pingserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/pingservice"
	vipserviceServer "oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/server/vipservice"
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
		// 健康检查服务
		crmrpc.RegisterPingServiceServer(grpcServer, pingserviceServer.NewPingServiceServer(ctx))

		// CRM统一初始化服务
		crmrpc.RegisterInitServer(grpcServer, initServer.NewInitServer(ctx))

		// 代理服务
		crmrpc.RegisterAuthServiceServer(grpcServer, authserviceServer.NewAuthServiceServer(ctx))
		crmrpc.RegisterAgentServiceServer(grpcServer, agentserviceServer.NewAgentServiceServer(ctx))

		// 会员自动层级服务
		crmrpc.RegisterMemberAutoLevelServer(grpcServer, memberautolevelServer.NewMemberAutoLevelServer(ctx))

		// VIP管理服务
		crmrpc.RegisterVipServiceServer(grpcServer, vipserviceServer.NewVipServiceServer(ctx))

		// 开发和测试环境额外开启服务反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
