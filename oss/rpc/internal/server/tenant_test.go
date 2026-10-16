package server

import (
	"context"
	"testing"

	"oa.98ent.com/p9/common/ctxdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRequireTenantFromClaims(t *testing.T) {
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op-a"})
	got, err := requireTenant(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c := ctxdata.ClaimsFromCtx(got)
	if c == nil || c.OperatorCode != "op-a" {
		t.Fatalf("claims: %+v", c)
	}
}

func TestRequireTenantFromMetadata(t *testing.T) {
	md := metadata.Pairs(headerOperatorCode, "op-b")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	got, err := requireTenant(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c := ctxdata.ClaimsFromCtx(got)
	if c == nil || c.OperatorCode != "op-b" {
		t.Fatalf("claims: %+v", c)
	}
}

func TestRequireTenantMissing(t *testing.T) {
	_, err := requireTenant(context.Background())
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		t.Fatalf("status: %v", err)
	}
}
