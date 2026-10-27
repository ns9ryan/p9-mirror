package middleware

import (
	"net/http"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/utils"
)

// ClientMeta 将 IP / User-Agent / X-Udid 写入上下文，供下游 CRM RPC 读取。
func ClientMeta(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = ctxdata.WithClientIP(ctx, utils.ClientIP(r))
		ctx = ctxdata.WithUserAgent(ctx, r.UserAgent())
		ctx = ctxdata.WithUdid(ctx, r.Header.Get("X-Udid"))
		next(w, r.WithContext(ctx))
	}
}
