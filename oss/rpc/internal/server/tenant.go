package server

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const headerOperatorCode = "x-operator-code"

func requireTenant(ctx context.Context) (context.Context, error) {
	code := tenantCode(ctx)
	if code == "" {
		return nil, status.Error(codes.Unauthenticated, "missing operator_code")
	}
	return ensureClaims(ctx, code), nil
}

func tenantCode(ctx context.Context) string {
	if s := strings.TrimSpace(ctxdata.OperatorCodeFromCtx(ctx)); s != "" {
		return s
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get(headerOperatorCode)
	if len(vals) == 0 {
		return ""
	}
	return strings.TrimSpace(vals[0])
}

func ensureClaims(ctx context.Context, code string) context.Context {
	if c := ctxdata.ClaimsFromCtx(ctx); c != nil {
		if strings.TrimSpace(c.OperatorCode) == code {
			return ctx
		}
		cp := *c
		cp.OperatorCode = code
		return ctxdata.WithClaims(ctx, &cp)
	}
	return ctxdata.WithClaims(ctx, &ctxdata.Claims{OperatorCode: code})
}
