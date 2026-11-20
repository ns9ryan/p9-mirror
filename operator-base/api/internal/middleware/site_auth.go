package middleware

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	coremiddleware "oa.98ent.com/p9/core/common/middleware"
)

// NewSiteAuth 复用 Core 登录及权限校验，限制具有站点身份的登录用户访问网站配置接口。
func NewSiteAuth(auth coremiddleware.Client) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		if auth == nil {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "no-store")
				httpx.ErrorCtx(r.Context(), w, xerr.Err(http.StatusServiceUnavailable, "core RPC is not configured"))
			}
		}
		// 此处只组装中间件；请求到来后依次校验 Token、站点身份及接口权限。
		protected := coremiddleware.JWT(auth)(requireSite(coremiddleware.Authority(auth)(next)))
		return func(w http.ResponseWriter, r *http.Request) {
			// 认证结果及站点配置响应不得缓存，避免复用其他登录身份的响应。
			w.Header().Set("Cache-Control", "no-store")
			if strings.TrimSpace(r.Header.Get("Authorization")) == "" {
				httpx.ErrorCtx(r.Context(), w, xerr.Unauthorized("authorization is required"))
				return
			}
			protected(w, r)
		}
	}
}

// requireSite 检查认证上下文是否包含站点编码；业务 Logic 仅按该编码查询和修改本站资料。
func requireSite(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := ctxdata.ClaimsFromCtx(r.Context())
		if claims == nil || strings.TrimSpace(claims.OperatorCode) == "" {
			httpx.ErrorCtx(r.Context(), w, xerr.Forbidden("login operator is required"))
			return
		}
		next(w, r)
	}
}
