// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/agent-admin-api/internal/logic/public"
	"oa.98ent.com/p9/agent-admin-api/internal/svc"
	"oa.98ent.com/p9/agent-admin-api/internal/types"
)

// 已开启的语言列表
func GetEnabledI18nLangsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetEnabledI18nLangsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := public.NewGetEnabledI18nLangsLogic(r.Context(), svcCtx)
		resp, err := l.GetEnabledI18nLangs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
