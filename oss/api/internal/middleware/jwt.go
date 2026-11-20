package middleware

import (
	"net/http"

	"oa.98ent.com/p9/core/common/coreadapt"
	coremw "oa.98ent.com/p9/core/common/middleware"
	"oa.98ent.com/p9/core/rpc/coreclient"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

const defaultCoreRPCTimeoutMS int64 = 10000

func passthrough(next http.HandlerFunc) http.HandlerFunc { return next }

// ClientIP 把客户端 IP 写入上下文（含 gRPC outgoing metadata）。
func ClientIP(next http.HandlerFunc) http.HandlerFunc {
	return coremw.ClientIP(next)
}

// JWT 校验 Authorization Bearer。
func JWT(c coremw.Client) rest.Middleware {
	return coremw.JWT(c)
}

// Authority Casbin Enforce（需在 JWT 之后）。
func Authority(c coremw.Client) rest.Middleware {
	return coremw.Authority(c)
}

// AuthFromRPC 用 core RPC 构造 JWT + Authority；endpoints 为空则均放行。
func AuthFromRPC(endpoints []string, timeoutMS int64) (coreclient.Core, rest.Middleware, rest.Middleware) {
	if len(endpoints) == 0 {
		return nil, passthrough, passthrough
	}
	if timeoutMS <= 0 {
		timeoutMS = defaultCoreRPCTimeoutMS
	}
	cli := coreclient.NewCore(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
		Timeout:   timeoutMS,
	}))
	auth := coreadapt.Auth(cli)
	return cli, JWT(auth), Authority(auth)
}

// JWTFromRPC 兼容旧调用：仅返回 JWT 中间件。
func JWTFromRPC(endpoints []string, timeoutMS int64) rest.Middleware {
	_, jwt, _ := AuthFromRPC(endpoints, timeoutMS)
	return jwt
}
