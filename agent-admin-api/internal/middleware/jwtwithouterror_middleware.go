// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package middleware

import (
	"github.com/zeromicro/go-zero/rest"
	"oa.98ent.com/p9/core/common/middleware"
)

// NewJwtWithoutErrorMiddleware 创建JWTWithoutError中间件
func NewJwtWithoutErrorMiddleware(cli middleware.Client) rest.Middleware {
	return middleware.JWTWithoutError(cli)
}
