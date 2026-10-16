package server

import (
	"oa.98ent.com/p9/oss/rpc/internal/config"
	"oa.98ent.com/p9/oss/rpc/internal/svc"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewZrpcServer(c config.Config, svcCtx *svc.ServiceContext) *zrpc.RpcServer {
	return zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterOssServer(grpcServer, NewOssServer(svcCtx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
}
