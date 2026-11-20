package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/config"
	memberWalletServer "oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/server/memberwallet"
	pingServer "oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/server/ping"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"
)

var configFile = flag.String("f", "etc/member_wallet.yaml", "the config file")

const (
	walletShutdownTimeout     = 30 * time.Second  // 单次钱包运行时关闭等待时间
	walletShutdownRetryDelay  = 1 * time.Second   // 关闭失败后的重试间隔
	walletShutdownMaxAttempts = 3                 // 最大关闭尝试次数
	processForceQuitTime      = 180 * time.Second // 进程优雅关闭最长等待时间
)

// closeServiceContext 负责关闭服务上下文, 必要时重试
func closeServiceContext(svcCtx *svc.ServiceContext) {
	for attempt := 1; attempt <= walletShutdownMaxAttempts; attempt++ {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), walletShutdownTimeout)

		err := svcCtx.Close(shutdownCtx)
		cancel()

		if err == nil {
			logx.Info("member wallet service context closed")
			return
		}

		// 只有等待超时或 Worker Pool 仍有任务时才重试, Actor 已完全退出但返回运行错误时, 不需要再次等待
		retryable := errors.Is(err, context.DeadlineExceeded) || errors.Is(err, wallet.ErrWorkerPoolBusy)

		if !retryable {
			logx.Errorf("close member wallet service context: %v", err)
			return
		}

		if attempt == walletShutdownMaxAttempts {
			logx.Errorf("member wallet shutdown incomplete after %d attempts: %v", attempt, err)
			return
		}

		logx.Errorf("member wallet shutdown attempt %d/%d failed, retrying: %v", attempt, walletShutdownMaxAttempts, err)

		time.Sleep(walletShutdownRetryDelay)
	}
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 为RPC优雅退出和钱包运行时排空预留时间
	proc.SetTimeToForceQuit(processForceQuitTime)

	// 初始化服务上下文
	svcCtx, err := svc.NewServiceContext(c)
	logx.Must(err)

	var s *zrpc.RpcServer

	// 先关闭钱包运行时和数据库, 最后关闭日志
	defer func() {
		closeServiceContext(svcCtx)

		if s != nil {
			s.Stop()
		}
	}()

	s = zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 健康检查服务
		memberwalletrpc.RegisterPingServer(grpcServer, pingServer.NewPingServer(svcCtx))

		// 会员钱包资金服务
		memberwalletrpc.RegisterMemberWalletServer(grpcServer, memberWalletServer.NewMemberWalletServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
