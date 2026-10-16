package middleware

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/client/authservice"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/authpb"
)

// jwtClient jwt认证客户端
type jwtClient struct {
	cli authservice.AuthService
}

// JwtClient 创建jwt认证客户端
func JwtClient(cli authservice.AuthService) middleware.Client {
	return &jwtClient{cli: cli}
}

// CheckToken 检查令牌
func (a *jwtClient) CheckToken(ctx context.Context, token string) (*ctxdata.Claims, error) {
	resp, err := a.cli.CheckToken(ctx, &authpb.CheckTokenReq{AccessToken: token})
	if err != nil {
		return nil, err
	}
	return &ctxdata.Claims{
		UserID:       resp.GetUserId(),
		Username:     resp.GetAccount(),
		OperatorCode: resp.GetOperatorCode(),
		Salt:         resp.GetSalt(),
		ExpiresAt:    resp.GetExpiresAt(),
		TokenType:    resp.GetTokenType(),
		ClientIP:     ctxdata.ClientIPFromCtx(ctx),
	}, nil

}

// Enforce 强制执行检查权限
func (a *jwtClient) Enforce(ctx context.Context, claims *ctxdata.Claims, path, method string) (bool, error) {
	return true, nil
}
