package nicknameservicelogic

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/platform-base/rpc/ent/nickname"
	_ "oa.98ent.com/p9/platform-base/rpc/ent/runtime"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/nicknamepb"
	"strings"
	"testing"
)

func TestWriteAuthorization(t *testing.T) {
	for _, claims := range []*ctxdata.Claims{nil, {UserID: 1, IsPlatform: false}, {UserID: 1, IsPlatform: true, TokenType: ctxdata.TokenTypePreview}} {
		ctx := ctxdata.WithClaims(context.Background(), claims)
		s := &svc.ServiceContext{}
		_, a := NewCreateLogic(ctx, s).Create(&nicknamepb.CreateNicknameRequest{Name: "昵称"})
		_, b := NewUpdateLogic(ctx, s).Update(&nicknamepb.UpdateNicknameRequest{Id: 1, Name: "昵称"})
		_, c := NewDeleteLogic(ctx, s).Delete(&nicknamepb.DeleteNicknameRequest{Id: 1})
		for _, err := range []error{a, b, c} {
			if status.Code(err) != codes.PermissionDenied {
				t.Fatalf("expected denial: %v", err)
			}
		}
	}
}
func TestInputBoundaries(t *testing.T) {
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, IsPlatform: true})
	l := NewCreateLogic(ctx, &svc.ServiceContext{})
	for _, name := range []string{"", " ", strings.Repeat("中", 101)} {
		if _, err := l.Create(&nicknamepb.CreateNicknameRequest{Name: name}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("name validation: %v", err)
		}
	}
	// The Ent validator must agree with RPC and HTTP on Unicode character counts.
	if err := nickname.NameValidator(strings.Repeat("中", 100)); err != nil {
		t.Fatalf("100 Unicode characters rejected: %v", err)
	}
	if err := nickname.NameValidator(strings.Repeat("中", 101)); err == nil {
		t.Fatal("101 Unicode characters accepted")
	}
	if _, err := NewGetLogic(ctx, &svc.ServiceContext{}).Get(&nicknamepb.GetNicknameRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	if _, err := NewListLogic(ctx, &svc.ServiceContext{}).List(&nicknamepb.ListNicknamesRequest{Page: 1, PageSize: 101}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
}
