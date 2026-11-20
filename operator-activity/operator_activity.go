package main

import (
	"flag"
	"fmt"

	taskserviceServer "oa.98ent.com/p9/operator-activity/internal/server/taskservice"
	promotion "oa.98ent.com/p9/operator-activity/modules/activity"
	"oa.98ent.com/p9/operator-activity/modules/activity/consumer"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/seed"
	"oa.98ent.com/p9/operator-activity/modules/temporal"
	"oa.98ent.com/p9/operator-activity/modules/worker"

	"oa.98ent.com/p9/operator-activity/internal/config"
	activityserviceServer "oa.98ent.com/p9/operator-activity/internal/server/activityservice"
	pingserviceServer "oa.98ent.com/p9/operator-activity/internal/server/pingservice"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	taskconfig "oa.98ent.com/p9/operator-activity/task/config"
	tasksvc "oa.98ent.com/p9/operator-activity/task/svc"
)

var (
	configFile = flag.String("f", "etc/operator_activity.yaml", "the config file")
	seedData   = flag.Bool("seed", false, "seed promotion test data")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	defer ctx.DB.Close()
	// 加载测试数据
	if *seedData {
		if err := seed.Run(ctx.DB); err != nil {
			logx.Errorf("seed failed: %v", err)
		} else {
			logx.Info("activity seed data loaded successfully")
		}
	}

	// 是否启动活动引擎
	if c.EnableActivityEngine {
		// 初始化Temporal
		temporal.Init(c.TemporalInfo.Address, c.TemporalInfo.Namespace)

		// 初始化活动模块
		rt, err := promotion.InitWithOptions(ctx, promotion.InitOptions{
			RedisAddr: c.RedisAddr,
		}, c.KqBrokers...)
		if err != nil {
			logx.Must(err)
		}
		ctx.ActivityRuntime = rt
		// 启动Kafka消费者
		var kqQueues []queue.MessageQueue
		if len(c.KqBrokers) > 0 {
			kqQueues = consumer.StartKq(rt.Scheduler, c)
			defer consumer.StopKq(kqQueues)
		}
		// 启动工作流worker
		go worker.Init(ctx.DB, c.TemporalInfo.Address, c.TemporalInfo.Namespace)
		// 启动定时任务
		// go schedule.CreateScheduleFromConfig()
		fmt.Println("工作流worker启动成功")
	}

	if c.EnableTaskEngine {
		// 初始化任务服务上下文
		taskCtx := tasksvc.InitializeTaskServiceContext(taskconfig.Config{
			Name:      c.Name,
			KqBrokers: c.TaskConfig.KqBrokers,
			RedisConf: taskconfig.RedisConf{
				Address:  c.TaskConfig.RedisConf.Address,
				Password: c.TaskConfig.RedisConf.Password,
				DB:       c.TaskConfig.RedisConf.DB,
			},
			EnableTaskTimer: c.TaskConfig.EnableTaskTimer,
			OperatorBaseRpc: c.OperatorBaseRpc,
		}, ctx.DB)
		defer taskCtx.Close()
		ctx.RedisQueue = taskCtx.RedisQueue
		logx.Info("Task engine initialized successfully")
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		operatoractivityrpc.RegisterPingServiceServer(grpcServer, pingserviceServer.NewPingServiceServer(ctx))
		operatoractivityrpc.RegisterActivityServiceServer(grpcServer, activityserviceServer.NewActivityServiceServer(ctx))
		operatoractivityrpc.RegisterTaskServiceServer(grpcServer, taskserviceServer.NewTaskServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
