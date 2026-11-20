package geo

import (
	"context"
	"testing"

	"oa.98ent.com/p9/common/ctxdata"

	"google.golang.org/grpc/metadata"
)

// 空则 default
func TestAccountOperatorCode_EmptyUsesDefault(t *testing.T) {
	if got := AccountOperatorCode(context.Background()); got != DefaultOperatorCode {
		t.Fatalf("got %q", got)
	}
}

// 用登录态
func TestAccountOperatorCode_FromClaims(t *testing.T) {
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op-a"})
	if got := AccountOperatorCode(ctx); got != "op-a" {
		t.Fatalf("got %q", got)
	}
}

func TestAccountOperatorCode_incomingMetadata(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-operator-code", "op-md"))
	if got := AccountOperatorCode(ctx); got != "op-md" {
		t.Fatalf("got %q", got)
	}
}

// 请求体优先于登录态
func TestResolveOperatorCode_ExplicitWins(t *testing.T) {
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op-a"})
	if got := ResolveOperatorCode(ctx, "op-body"); got != "op-body" {
		t.Fatalf("got %q", got)
	}
	if got := AccountOperatorCode(ctx); got != "op-a" {
		t.Fatalf("account got %q", got)
	}
}
