package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"oa.98ent.com/p9/operator-report/rpc/internal/config"
	statisticsservice "oa.98ent.com/p9/operator-report/rpc/internal/server/statisticsservice"
	"oa.98ent.com/p9/operator-report/rpc/internal/svc"
	"oa.98ent.com/p9/operator-report/rpc/pb/operator_report"
)

var configFile = flag.String("f", "etc/operator_report.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.DB.Close()

	// // 初始化定时任务调度器
	// cronScheduler, err := cron.InitCron()
	// if err != nil {
	// 	logx.Errorf("初始化定时任务调度器失败: %v", err)
	// 	return
	// }
	// cronScheduler.Start()
	// defer cronScheduler.Stop()
	// logx.Infof("定时任务调度器已启动")

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// Register all individual services
		operator_report.RegisterStatisticsServiceServer(grpcServer, statisticsservice.NewStatisticsServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
