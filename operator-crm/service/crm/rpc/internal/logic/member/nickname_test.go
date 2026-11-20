package memberlogic

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/client/nicknameservice"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
	"strings"
	"testing"
)

type nicknameStub struct {
	nicknameservice.NicknameService
	calls int
	name  string
	err   error
}

func (s *nicknameStub) Random(context.Context, *nicknamepb.RandomNicknameRequest, ...grpc.CallOption) (*nicknamepb.RandomNicknameResponse, error) {
	s.calls++
	return &nicknamepb.RandomNicknameResponse{Nickname: &nicknamepb.NicknameInfo{Name: s.name}}, s.err
}
func TestResolveNickname(t *testing.T) {
	for _, tc := range []struct {
		name, input, pool string
		rpcErr            error
		want              string
		code              codes.Code
		calls             int
	}{
		{name: "explicit", input: " 自选昵称 ", want: "自选昵称"},
		{name: "random", input: "  ", pool: "随机昵称", want: "随机昵称", calls: 1},
		{name: "empty pool", rpcErr: status.Error(codes.NotFound, "empty"), want: "account", calls: 1},
		{name: "RPC failure", rpcErr: status.Error(codes.Unavailable, "down"), code: codes.Unavailable, calls: 1},
		{name: "invalid pool", code: codes.Internal, calls: 1},
		{name: "too long", input: strings.Repeat("中", 101), code: codes.InvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &nicknameStub{name: tc.pool, err: tc.rpcErr}
			l := NewCreateLogic(context.Background(), &svc.ServiceContext{Nicknames: stub})
			got, err := l.resolveNickname(tc.input, "account")
			if got != tc.want || status.Code(err) != tc.code || stub.calls != tc.calls {
				t.Fatalf("got %q %v calls=%d", got, err, stub.calls)
			}
		})
	}
	l := NewCreateLogic(context.Background(), &svc.ServiceContext{})
	if _, err := l.resolveNickname("", "account"); status.Code(err) != codes.Unavailable {
		t.Fatalf("missing config: %v", err)
	}
}
