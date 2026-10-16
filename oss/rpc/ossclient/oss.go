// Package ossclient 是 oss 对外 gRPC 客户端（档案 / 服务商）。
// 调用方传入 HTTP JWT 后的 ctx 即可；登录态已在 ctxdata.WithClaims 写入 outgoing metadata。
package ossclient

import (
	"context"

	oss "oa.98ent.com/p9/oss/rpc/types/oss"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	Empty         = oss.Empty
	IDReq         = oss.IDReq
	KeyReq        = oss.KeyReq
	BucketSpec    = oss.BucketSpec
	SceneList     = oss.SceneList
	Provider      = oss.Provider
	ProviderList  = oss.ProviderList
	FileData      = oss.FileData
	FileInfo      = oss.FileInfo
	FileListReq   = oss.FileListReq
	FileList      = oss.FileList
	FileRecordReq = oss.FileRecordReq
	InitTenantReq = oss.InitTenantReq
)

type Oss interface {
	InitTenant(ctx context.Context, in *InitTenantReq, opts ...grpc.CallOption) (*Provider, error)
	CreateProvider(ctx context.Context, in *Provider, opts ...grpc.CallOption) (*Provider, error)
	UpdateProvider(ctx context.Context, in *Provider, opts ...grpc.CallOption) (*Provider, error)
	ListProviders(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ProviderList, error)
	GetProvider(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*Provider, error)
	DeleteProvider(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*Empty, error)
	EnableProvider(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*Provider, error)
	GetActiveProvider(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Provider, error)
	GetStorageConfig(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Provider, error)
	ListFiles(ctx context.Context, in *FileListReq, opts ...grpc.CallOption) (*FileList, error)
	DeleteFile(ctx context.Context, in *KeyReq, opts ...grpc.CallOption) (*FileData, error)
	RecordFile(ctx context.Context, in *FileRecordReq, opts ...grpc.CallOption) (*Empty, error)
}

type defaultOss struct {
	cli zrpc.Client
}

func NewOss(cli zrpc.Client) Oss {
	return &defaultOss{cli: cli}
}

func (m *defaultOss) conn() oss.OssClient {
	return oss.NewOssClient(m.cli.Conn())
}

func (m *defaultOss) InitTenant(ctx context.Context, in *InitTenantReq, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().InitTenant(ctx, in, opts...)
}

func (m *defaultOss) CreateProvider(ctx context.Context, in *Provider, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().CreateProvider(ctx, in, opts...)
}

func (m *defaultOss) UpdateProvider(ctx context.Context, in *Provider, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().UpdateProvider(ctx, in, opts...)
}

func (m *defaultOss) ListProviders(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ProviderList, error) {
	return m.conn().ListProviders(ctx, in, opts...)
}

func (m *defaultOss) GetProvider(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().GetProvider(ctx, in, opts...)
}

func (m *defaultOss) DeleteProvider(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*Empty, error) {
	return m.conn().DeleteProvider(ctx, in, opts...)
}

func (m *defaultOss) EnableProvider(ctx context.Context, in *IDReq, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().EnableProvider(ctx, in, opts...)
}

func (m *defaultOss) GetActiveProvider(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().GetActiveProvider(ctx, in, opts...)
}

func (m *defaultOss) GetStorageConfig(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Provider, error) {
	return m.conn().GetStorageConfig(ctx, in, opts...)
}

func (m *defaultOss) ListFiles(ctx context.Context, in *FileListReq, opts ...grpc.CallOption) (*FileList, error) {
	return m.conn().ListFiles(ctx, in, opts...)
}

func (m *defaultOss) DeleteFile(ctx context.Context, in *KeyReq, opts ...grpc.CallOption) (*FileData, error) {
	return m.conn().DeleteFile(ctx, in, opts...)
}

func (m *defaultOss) RecordFile(ctx context.Context, in *FileRecordReq, opts ...grpc.CallOption) (*Empty, error) {
	return m.conn().RecordFile(ctx, in, opts...)
}
