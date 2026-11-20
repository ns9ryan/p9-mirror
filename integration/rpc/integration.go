package main

import (
	"flag"
	"fmt"

	"oa.98ent.com/p9/integration/rpc/internal/config"
	"oa.98ent.com/p9/integration/rpc/internal/server"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "rpc/etc/integration.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	s := zrpc.MustNewServer(messageRPCConfig(c.RpcServerConf), func(grpcServer *grpc.Server) {
		integration.RegisterIntegrationServer(grpcServer, server.NewIntegrationServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddOptions(
		grpc.MaxRecvMsgSize(16<<20),
		grpc.MaxSendMsgSize(16<<20),
	)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// messageRPCConfig 禁止消息接口记录请求正文，避免验证码、密送地址和附件进入访问日志。
// 保留其他方法及调用方已有的日志配置；Log.Stat 只控制统计日志，不控制请求正文。
func messageRPCConfig(c zrpc.RpcServerConf) zrpc.RpcServerConf {
	ignored := append([]string(nil), c.Middlewares.StatConf.IgnoreContentMethods...)
	c.Middlewares.StatConf.IgnoreContentMethods = append(ignored,
		integration.Integration_SendEmail_FullMethodName,
		integration.Integration_SendSms_FullMethodName,
	)
	return c
}
