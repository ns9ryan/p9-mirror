package ctxdata

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

// ctxKey 上下文键
type ctxKey int

const (
	claimsKey ctxKey = iota + 1
	rawTokenKey
	clientIPKey
	userAgentKey
	udidKey
)

// header 头部
const (
	headerUserID       = "x-user-id"
	headerUserCode     = "x-user-code"
	headerUsername     = "x-username"
	headerUserType     = "x-user-type"
	headerOperatorCode = "x-operator-code"
	headerRoleCodes    = "x-role-codes"
	headerSalt         = "x-salt"
	headerExpiresAt    = "x-expires-at"
	headerRawToken     = "x-raw-token"
	headerIsPlatform   = "x-is-platform"
	headerTokenType    = "x-token-type"
	headerClientIP     = "x-client-ip"
	headerUserAgent    = "x-user-agent"
	headerUdid         = "x-udid"
)

// skipTenantKey 跳过租户
type skipTenantKey struct{}

// skipSoftDeleteKey 跳过软删除
type skipSoftDeleteKey struct{}

// claimsBox 身份盒子
type claimsBox struct {
	c *Claims
}

// Claims 用户身份信息
type Claims struct {
	UserID       int64     // 用户ID
	UserCode     string    // 用户代码
	Username     string    // 用户名
	UserType     UserType  // 用户类型 [platform, operator, agent]
	OperatorCode string    // 操作员代码
	RoleCodes    []string  // 角色代码
	Salt         string    // 盐
	ExpiresAt    int64     // 过期时间
	IsPlatform   bool      // 是否是总网
	TokenType    TokenType // 令牌类型 [access, refresh, preview]
	ClientIP     string    // 客户端IP
}

// WithClaimsHolder 在外层放入可写的身份盒子，内层 WithClaims 会回写，外层仍能读到。
func WithClaimsHolder(ctx context.Context) context.Context {
	if boxFrom(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, claimsKey, &claimsBox{})
}

// WithClaims 添加用户身份信息到上下文
func WithClaims(ctx context.Context, c *Claims) context.Context {
	if c == nil {
		return ctx
	}
	if box := boxFrom(ctx); box != nil {
		box.c = c
	} else {
		ctx = context.WithValue(ctx, claimsKey, &claimsBox{c: c})
	}
	return metadata.AppendToOutgoingContext(ctx,
		headerUserID, strconv.FormatInt(c.UserID, 10),
		headerUserCode, c.UserCode,
		headerUsername, c.Username,
		headerUserType, string(c.UserType),
		headerOperatorCode, c.OperatorCode,
		headerRoleCodes, strings.Join(c.RoleCodes, ","),
		headerSalt, c.Salt,
		headerExpiresAt, strconv.FormatInt(c.ExpiresAt, 10),
		headerIsPlatform, formatBool(c.IsPlatform),
		headerTokenType, c.TokenType.String(),
		headerClientIP, c.ClientIP,
	)
}

// ClaimsFromCtx 从上下文中获取用户身份信息
func ClaimsFromCtx(ctx context.Context) *Claims {
	if box := boxFrom(ctx); box != nil && box.c != nil {
		return box.c
	}
	return claimsFromIncomingMD(ctx)
}

func boxFrom(ctx context.Context) *claimsBox {
	box, _ := ctx.Value(claimsKey).(*claimsBox)
	return box
}

// OperatorCodeFromCtx 从上下文中获取操作员代码
func OperatorCodeFromCtx(ctx context.Context) string {
	c := ClaimsFromCtx(ctx)
	if c == nil {
		return ""
	}
	return c.OperatorCode
}

// UserIDFromCtx 从上下文中获取用户ID
func UserIDFromCtx(ctx context.Context) int64 {
	c := ClaimsFromCtx(ctx)
	if c == nil {
		return 0
	}
	return c.UserID
}

// WithOperatorCode 添加分站代码到上下文
func WithOperatorCode(ctx context.Context, code string) context.Context {
	if c := ClaimsFromCtx(ctx); c != nil {
		next := *c
		next.OperatorCode = code
		return WithClaims(ctx, &next)
	}
	return WithClaims(ctx, &Claims{OperatorCode: code})
}

// WithRawToken 添加原始令牌到上下文
func WithRawToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	ctx = context.WithValue(ctx, rawTokenKey, token)
	return metadata.AppendToOutgoingContext(ctx, headerRawToken, token)
}

// WithUserType 添加用户类型到上下文
func WithUserType(ctx context.Context, userType UserType) context.Context {
	if c := ClaimsFromCtx(ctx); c != nil {
		next := *c
		next.UserType = userType
		return WithClaims(ctx, &next)
	}
	return WithClaims(ctx, &Claims{UserType: userType})
}

// UserTypeFromCtx 从上下文中获取用户类型
func UserTypeFromCtx(ctx context.Context) UserType {
	c := ClaimsFromCtx(ctx)
	if c == nil {
		return ""
	}
	return c.UserType
}

// RawTokenFromCtx 从上下文中获取原始令牌
func RawTokenFromCtx(ctx context.Context) string {
	if s, _ := ctx.Value(rawTokenKey).(string); s != "" {
		return s
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return first(md, headerRawToken)
}

// WithClientIP 添加客户端 IP 到上下文，并写入 gRPC outgoing metadata。
func WithClientIP(ctx context.Context, ip string) context.Context {
	if ip == "" {
		return ctx
	}
	ctx = context.WithValue(ctx, clientIPKey, ip)
	return metadata.AppendToOutgoingContext(ctx, headerClientIP, ip)
}

// ClientIPFromCtx 从上下文中获取客户端 IP。
func ClientIPFromCtx(ctx context.Context) string {
	if s, _ := ctx.Value(clientIPKey).(string); s != "" {
		return s
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return first(md, headerClientIP)
}

// WithUserAgent 添加 User-Agent 到上下文，并写入 gRPC outgoing metadata。
func WithUserAgent(ctx context.Context, ua string) context.Context {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return ctx
	}
	ctx = context.WithValue(ctx, userAgentKey, ua)
	return metadata.AppendToOutgoingContext(ctx, headerUserAgent, ua)
}

// UserAgentFromCtx 从上下文中获取 User-Agent。
func UserAgentFromCtx(ctx context.Context) string {
	if s, _ := ctx.Value(userAgentKey).(string); s != "" {
		return s
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return first(md, headerUserAgent)
}

// WithUdid 添加设备码到上下文，并写入 gRPC outgoing metadata。
func WithUdid(ctx context.Context, udid string) context.Context {
	udid = strings.TrimSpace(udid)
	if udid == "" {
		return ctx
	}
	ctx = context.WithValue(ctx, udidKey, udid)
	return metadata.AppendToOutgoingContext(ctx, headerUdid, udid)
}

// UdidFromCtx 从上下文中获取设备码。
func UdidFromCtx(ctx context.Context) string {
	if s, _ := ctx.Value(udidKey).(string); s != "" {
		return s
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return first(md, headerUdid)
}

// SkipTenant 跳过租户
func SkipTenant(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipTenantKey{}, true)
}

// TenantSkipped 是否跳过租户
func TenantSkipped(ctx context.Context) bool {
	ok, _ := ctx.Value(skipTenantKey{}).(bool)
	return ok
}

// SkipSoftDelete 跳过软删除
func SkipSoftDelete(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipSoftDeleteKey{}, true)
}

// SoftDeleteSkipped 是否跳过软删除
func SoftDeleteSkipped(ctx context.Context) bool {
	ok, _ := ctx.Value(skipSoftDeleteKey{}).(bool)
	return ok
}

// claimsFromIncomingMD 从入站元数据中获取用户身份信息
func claimsFromIncomingMD(ctx context.Context) *Claims {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	userID := parseInt(first(md, headerUserID))
	username := first(md, headerUsername)
	userType := UserType(first(md, headerUserType))
	operatorCode := first(md, headerOperatorCode)
	isPlatform := parseBool(first(md, headerIsPlatform))
	if userID == 0 && username == "" && operatorCode == "" && !isPlatform {
		return nil
	}
	codes := []string{}
	if raw := first(md, headerRoleCodes); raw != "" {
		codes = strings.Split(raw, ",")
	}
	return &Claims{
		UserID:       userID,
		UserCode:     first(md, headerUserCode),
		Username:     username,
		UserType:     userType,
		OperatorCode: operatorCode,
		RoleCodes:    codes,
		Salt:         first(md, headerSalt),
		ExpiresAt:    parseInt(first(md, headerExpiresAt)),
		IsPlatform:   isPlatform,
		TokenType:    TokenType(first(md, headerTokenType)),
		ClientIP:     first(md, headerClientIP),
	}
}

// first 获取第一个值
func first(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// parseInt 解析整数
func parseInt(s string) int64 {
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// formatBool 格式化布尔值
func formatBool(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

// parseBool 解析布尔值
func parseBool(s string) bool {
	return s == "1" || strings.EqualFold(s, "true")
}
