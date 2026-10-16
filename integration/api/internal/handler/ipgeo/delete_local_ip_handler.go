package ipgeo

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/integration/api/internal/logic/ipgeo"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
)

// DeleteLocalIPHandler 删除本地 IP。
func DeleteLocalIPHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IDReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ipgeo.NewDeleteLocalIPLogic(r.Context(), svcCtx)
		resp, err := l.DeleteLocalIP(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
