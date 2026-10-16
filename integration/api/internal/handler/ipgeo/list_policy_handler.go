package ipgeo

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/integration/api/internal/logic/ipgeo"
	"oa.98ent.com/p9/integration/api/internal/svc"
)

// ListPolicyHandler 当前租户策略列表。
func ListPolicyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ipgeo.NewListPolicyLogic(r.Context(), svcCtx)
		resp, err := l.ListPolicy()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
