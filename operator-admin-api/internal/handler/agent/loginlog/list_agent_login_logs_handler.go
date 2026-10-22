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

// 代理登录日志列表
func ListAgentLoginLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAgentLoginLogsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := loginlog.NewListAgentLoginLogsLogic(r.Context(), svcCtx)
		resp, err := l.ListAgentLoginLogs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
