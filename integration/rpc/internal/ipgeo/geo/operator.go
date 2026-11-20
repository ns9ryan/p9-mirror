package geo

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

// DefaultOperatorCode 未指定租户时的默认 operator_code。
const DefaultOperatorCode = "default"

// NormalizeOperatorCode 去空白，空则返回 default。
func NormalizeOperatorCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return DefaultOperatorCode
	}
	return code
}

// ResolveOperatorCode 请求体 > 登录态 / incoming metadata > default。
func ResolveOperatorCode(ctx context.Context, explicit string) string {
	code := resolveOperatorCode(ctx, explicit)
	logRPCLogin(ctx, explicit, code)
	return code
}

// AccountOperatorCode 只用登录态 / incoming metadata，忽略请求体。
func AccountOperatorCode(ctx context.Context) string {
	return ResolveOperatorCode(ctx, "")
}

func resolveOperatorCode(ctx context.Context, explicit string) string {
	if s := strings.TrimSpace(explicit); s != "" {
		return s
	}
	if s := strings.TrimSpace(ctxdata.OperatorCodeFromCtx(ctx)); s != "" {
		return s
	}
	if s := operatorCodeFromIncomingMD(ctx); s != "" {
		return s
	}
	return DefaultOperatorCode
}

func operatorCodeFromIncomingMD(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("x-operator-code")
	if len(vals) == 0 {
		return ""
	}
	return strings.TrimSpace(vals[0])
}

func logRPCLogin(ctx context.Context, explicit, resolved string) {
	c := ctxdata.ClaimsFromCtx(ctx)
	l := logx.WithContext(ctx)
	if c == nil {
		l.Infof("rpc login empty explicit=%q resolved=%s", explicit, resolved)
		return
	}
	l.Infof("rpc login user_id=%d username=%s operator_code=%s",
		c.UserID, c.Username, c.OperatorCode)
	l.Infof("rpc login is_platform=%v token_type=%s explicit=%q resolved=%s",
		c.IsPlatform, c.TokenType, explicit, resolved)
}
