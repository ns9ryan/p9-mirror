// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package level

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/agent/level"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
)

// 代理等级列表
func ListAgentLevelsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := level.NewListAgentLevelsLogic(r.Context(), svcCtx)
		resp, err := l.ListAgentLevels()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
