package memberexport

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
)

func TestCellSafetyAndFidelity(t *testing.T) {
	for _, v := range []string{"=SUM(A1)", "+1", "-1", "@x", "\t123", "\r123", "\n123", "  =cmd"} {
		if got := Cell(v, false); got != "'"+v {
			t.Fatalf("unprotected %q", v)
		}
	}
	for _, v := range []string{"001234567890", "9223372036854775807", "0013800000000"} {
		if Cell(v, true) != "'"+v {
			t.Fatal("identifier changed")
		}
	}
	if Cell("9007199254740993", false) != "9007199254740993" {
		t.Fatal("amount rounded")
	}
	if Cell(Missing, true) != Missing {
		t.Fatal("missing value changed")
	}
}

type fakeAuth struct {
	claims *ctxdata.Claims
	deny   string
	fail   string
	calls  int
}

type forwardingAuth struct {
	fakeAuth
	check func(context.Context)
}

func (f *forwardingAuth) CheckToken(ctx context.Context, _ string) (*ctxdata.Claims, error) {
	f.check(ctx)
	return f.claims, nil
}

func TestExportForwardsClientIPToCore(t *testing.T) {
	claims := &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "OP_test", TokenType: ctxdata.TokenTypeAccess, ClientIP: "127.0.0.1"}
	origin := ctxdata.WithClientIP(ctxdata.WithRawToken(ctxdata.WithClaims(context.Background(), claims), "test-token"), "127.0.0.1")
	md, _ := metadata.FromOutgoingContext(origin)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	auth := &forwardingAuth{fakeAuth: fakeAuth{claims: claims}, check: func(ctx context.Context) {
		out, _ := metadata.FromOutgoingContext(ctx)
		coreCtx := metadata.NewIncomingContext(context.Background(), out)
		if got := ctxdata.ClientIPFromCtx(coreCtx); got != "127.0.0.1" {
			t.Fatalf("Core received client IP %q", got)
		}
	}}
	if _, _, err := Authorize(ctx, auth); err != nil {
		t.Fatal(err)
	}
}

func (f *fakeAuth) CheckToken(context.Context, string) (*ctxdata.Claims, error) { return f.claims, nil }
func (f *fakeAuth) Enforce(_ context.Context, _ *ctxdata.Claims, path, method string) (bool, error) {
	f.calls++
	if f.fail == path {
		return false, errors.New("dependency failure")
	}
	return f.deny != path, nil
}
func TestExportPolicy(t *testing.T) {
	claims := &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "OP_test", TokenType: ctxdata.TokenTypeAccess}
	ctx := ctxdata.WithRawToken(ctxdata.WithClaims(context.Background(), claims), "test-token")
	auth := &fakeAuth{claims: claims}
	_, p, err := Authorize(ctx, auth)
	if err != nil || p.Phone || p.Email || auth.calls != 1 {
		t.Fatalf("policy not resolved once: %+v %v", p, err)
	}
	auth.fail = Path
	if _, _, err = Authorize(ctx, auth); status.Code(err) != codes.Unavailable {
		t.Fatal("permission failure allowed")
	}
	auth.fail = ""
	auth.deny = Path
	if _, _, err = Authorize(ctx, auth); status.Code(err) != codes.PermissionDenied {
		t.Fatal("missing export permission allowed")
	}
	if _, _, err = Authorize(context.Background(), auth); err == nil {
		t.Fatal("missing credentials allowed")
	}
	other := *claims
	other.OperatorCode = "OP_other"
	auth.claims = &other
	if _, _, err = Authorize(ctx, auth); err == nil {
		t.Fatal("cross tenant allowed")
	}
	for _, kind := range []ctxdata.UserType{ctxdata.UserTypePlatform, ctxdata.UserTypeAgent} {
		other = *claims
		other.UserType = kind
		auth.claims = &other
		if _, _, err = Authorize(ctx, auth); err == nil {
			t.Fatal("non operator allowed")
		}
	}
}

func TestColumnSelection(t *testing.T) {
	indices, err := ColumnIndices([]string{"status", "username", "member_code"})
	if err != nil || len(indices) != 3 || indices[0] != 6 || indices[1] != 2 || indices[2] != 1 {
		t.Fatal("column order lost")
	}
	for _, columns := range [][]string{{"username", "username"}, {"password_hash"}, {"phone_ciphertext"}, {"unknown"}, {""}} {
		if _, err = ColumnIndices(columns); err == nil {
			t.Fatal("invalid or sensitive raw column accepted")
		}
	}
	if indices, err = ColumnIndices(nil); err != nil || len(indices) != len(Header) {
		t.Fatal("default columns changed")
	}
}
