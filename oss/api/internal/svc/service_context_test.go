package svc

import (
	"context"
	"testing"

	"oa.98ent.com/p9/oss/rpc/ossclient"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type storageCfgStub struct {
	ossclient.Oss
	p   *ossclient.Provider
	err error
}

func (s storageCfgStub) GetStorageConfig(
	_ context.Context, _ *ossclient.Empty, _ ...grpc.CallOption,
) (*ossclient.Provider, error) {
	return s.p, s.err
}

func TestStore_未配档案RPC用本地驱动(t *testing.T) {
	if _, _, err := (&ServiceContext{}).Store(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStore_无服务商回退本地oss(t *testing.T) {
	s := &ServiceContext{Oss: storageCfgStub{p: &ossclient.Provider{Name: "(none)", Source: "yaml"}}}
	if _, _, err := s.Store(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStore_GetStorageConfigNotFound回退本地oss(t *testing.T) {
	s := &ServiceContext{Oss: storageCfgStub{err: status.Error(codes.NotFound, "no active storage provider")}}
	if _, _, err := s.Store(context.Background()); err != nil {
		t.Fatal(err)
	}
}
