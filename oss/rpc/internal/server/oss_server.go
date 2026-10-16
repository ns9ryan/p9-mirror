package server

import (
	"context"

	filelogic "oa.98ent.com/p9/oss/rpc/internal/logic/file"
	providerlogic "oa.98ent.com/p9/oss/rpc/internal/logic/provider"
	"oa.98ent.com/p9/oss/rpc/internal/svc"
	pb "oa.98ent.com/p9/oss/rpc/types/oss"
)

type OssServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedOssServer
}

func NewOssServer(svcCtx *svc.ServiceContext) *OssServer {
	return &OssServer{svcCtx: svcCtx}
}

func (s *OssServer) InitTenant(ctx context.Context, in *pb.InitTenantReq) (*pb.Provider, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).InitTenant(in)
}

func (s *OssServer) CreateProvider(ctx context.Context, in *pb.Provider) (*pb.Provider, error) {
	ctx, err := requireTenant(ctx)
	if err != nil {
		return nil, err
	}
	return providerlogic.NewLogic(ctx, s.svcCtx).CreateProvider(in)
}

func (s *OssServer) UpdateProvider(ctx context.Context, in *pb.Provider) (*pb.Provider, error) {
	ctx, err := requireTenant(ctx)
	if err != nil {
		return nil, err
	}
	return providerlogic.NewLogic(ctx, s.svcCtx).UpdateProvider(in)
}

func (s *OssServer) ListProviders(ctx context.Context, in *pb.Empty) (*pb.ProviderList, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).ListProviders(in)
}

func (s *OssServer) GetProvider(ctx context.Context, in *pb.IDReq) (*pb.Provider, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).GetProvider(in)
}

func (s *OssServer) DeleteProvider(ctx context.Context, in *pb.IDReq) (*pb.Empty, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).DeleteProvider(in)
}

func (s *OssServer) EnableProvider(ctx context.Context, in *pb.IDReq) (*pb.Provider, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).EnableProvider(in)
}

func (s *OssServer) GetActiveProvider(ctx context.Context, in *pb.Empty) (*pb.Provider, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).GetActiveProvider(in)
}

func (s *OssServer) GetStorageConfig(ctx context.Context, in *pb.Empty) (*pb.Provider, error) {
	return providerlogic.NewLogic(ctx, s.svcCtx).GetStorageConfig(in)
}

func (s *OssServer) ListFiles(ctx context.Context, in *pb.FileListReq) (*pb.FileList, error) {
	return filelogic.NewListLogic(ctx, s.svcCtx).ListFiles(in)
}

func (s *OssServer) DeleteFile(ctx context.Context, in *pb.KeyReq) (*pb.FileData, error) {
	ctx, err := requireTenant(ctx)
	if err != nil {
		return nil, err
	}
	return filelogic.NewDeleteLogic(ctx, s.svcCtx).DeleteFile(in)
}

func (s *OssServer) RecordFile(ctx context.Context, in *pb.FileRecordReq) (*pb.Empty, error) {
	return filelogic.NewRecordLogic(ctx, s.svcCtx).RecordFile(in)
}
