// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package loginlog

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/agent/loginlog"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

// 关联检索登录日志
func ListAgentLoginAssociationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAgentLoginAssociationsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := loginlog.NewListAgentLoginAssociationsLogic(r.Context(), svcCtx)
		resp, err := l.ListAgentLoginAssociations(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
