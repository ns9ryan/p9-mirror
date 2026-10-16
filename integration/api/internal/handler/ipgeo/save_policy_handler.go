package ipgeo

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/integration/api/internal/logic/ipgeo"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
)

// SavePolicyHandler 保存当前租户策略。
func SavePolicyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SavePolicyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ipgeo.NewSavePolicyLogic(r.Context(), svcCtx)
		resp, err := l.SavePolicy(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
