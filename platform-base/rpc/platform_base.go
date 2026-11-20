package main

import (
	"flag"
	"fmt"
	nicknameserver "oa.98ent.com/p9/platform-base/rpc/internal/server/nicknameservice"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"oa.98ent.com/p9/platform-base/rpc/internal/config"
	currencyserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/currencyservice"
	defaultavatarframeserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/defaultavatarframeservice"
	defaultavatarserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/defaultavatarservice"
	pingserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/pingservice"
	regionserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/regionservice"
	timezoneserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/timezoneservice"
	vipdefaulticonserviceServer "oa.98ent.com/p9/platform-base/rpc/internal/server/vipdefaulticonservice"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc"
)

var configFile = flag.String("f", "etc/platform_base.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.DB.Close()

	// 执行数据库自动迁移
	ctx.MustMigrate()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		platformbaserpc.RegisterNicknameServiceServer(grpcServer, nicknameserver.NewNicknameServiceServer(ctx))
		// 默认头像服务，供总站管理和分站读取。
		platformbaserpc.RegisterDefaultAvatarServiceServer(grpcServer, defaultavatarserviceServer.NewDefaultAvatarServiceServer(ctx))
		// 默认头像框服务，供总站管理和分站读取。
		platformbaserpc.RegisterDefaultAvatarFrameServiceServer(grpcServer, defaultavatarframeserviceServer.NewDefaultAvatarFrameServiceServer(ctx))
		// Ping 服务
		platformbaserpc.RegisterPingServiceServer(grpcServer, pingserviceServer.NewPingServiceServer(ctx))

		// 时区服务
		platformbaserpc.RegisterTimezoneServiceServer(grpcServer, timezoneserviceServer.NewTimezoneServiceServer(ctx))

		// 货币服务
		platformbaserpc.RegisterCurrencyServiceServer(grpcServer, currencyserviceServer.NewCurrencyServiceServer(ctx))

		// 国家地区服务
		platformbaserpc.RegisterRegionServiceServer(grpcServer, regionserviceServer.NewRegionServiceServer(ctx))

		// 默认 VIP 图标服务
		platformbaserpc.RegisterVipDefaultIconServiceServer(grpcServer, vipdefaulticonserviceServer.NewVipDefaultIconServiceServer(ctx))

		// 开发和测试环境额外开启服务反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
