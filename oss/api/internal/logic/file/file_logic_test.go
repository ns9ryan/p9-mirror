package file

import (
	"context"
	"testing"

	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/rpc/ossclient"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type deleteArchiveStub struct {
	ossclient.Oss
	err error
}

func (s deleteArchiveStub) DeleteFile(
	_ context.Context, _ *ossclient.KeyReq, _ ...grpc.CallOption,
) (*ossclient.FileData, error) {
	return nil, s.err
}

func TestDelete_本租户没有该档案则404(t *testing.T) {
	l := NewDeleteLogic(context.Background(), &svc.ServiceContext{
		Oss: deleteArchiveStub{err: status.Error(codes.NotFound, "file not found")},
	})
	_, err := l.Delete("public/20260101/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg")
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("err=%v", err)
	}
}

func TestDelete_空key(t *testing.T) {
	l := NewDeleteLogic(context.Background(), &svc.ServiceContext{})
	if _, err := l.Delete("  "); err == nil {
		t.Fatal("expected error")
	}
}
