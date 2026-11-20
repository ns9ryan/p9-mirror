package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
)

const (
	operatorAuthScopeHeader = "X-OSS-Auth-Scope"
	operatorAuthScope       = "operator"
	operatorUploadPath      = "/p9/oss/file/upload"
)

// OperatorUploadAuth 为分站上传构造独立鉴权链，不回退到总站 Core。
func OperatorUploadAuth(configured bool, jwt, actionLog, authority rest.Middleware) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		verifyIdentity := func(w http.ResponseWriter, r *http.Request) {
			claims := ctxdata.ClaimsFromCtx(r.Context())
			if claims == nil || claims.UserID <= 0 || !claims.UserType.IsOperator() || claims.OperatorCode == "" || claims.IsPlatform || claims.TokenType != ctxdata.TokenTypeAccess {
				httpx.ErrorCtx(r.Context(), w, xerr.Forbidden("operator access token is required"))
				return
			}
			actionLog(authority(next))(w, r)
		}
		authenticated := jwt(verifyIdentity)
		return func(w http.ResponseWriter, r *http.Request) {
			if !configured {
				httpx.ErrorCtx(r.Context(), w, xerr.Err(http.StatusServiceUnavailable, "operator_core_rpc is not configured"))
				return
			}
			authenticated(w, r)
		}
	}
}

// SelectUploadAuth 仅允许明确标识的上传请求使用分站鉴权，其余请求保持原鉴权链。
// 请求头只选择验证方，不提供身份；身份和权限必须由对应 Core 验证。
func SelectUploadAuth(defaultAuth, operatorAuth rest.Middleware) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		defaultHandler := defaultAuth(next)
		operatorHandler := operatorAuth(next)
		return func(w http.ResponseWriter, r *http.Request) {
			scope := r.Header.Get(operatorAuthScopeHeader)
			if scope == "" {
				defaultHandler(w, r)
				return
			}
			if scope != operatorAuthScope || r.Method != http.MethodPost || r.URL.Path != operatorUploadPath {
				httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("unsupported OSS authentication scope"))
				return
			}
			operatorHandler(w, r)
		}
	}
}
