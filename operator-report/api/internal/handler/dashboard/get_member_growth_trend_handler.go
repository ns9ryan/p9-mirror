// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dashboard

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-report/api/internal/logic/dashboard"
	"oa.98ent.com/p9/operator-report/api/internal/svc"
	"oa.98ent.com/p9/operator-report/api/internal/types"
)

func GetMemberGrowthTrendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MemberGrowthTrendRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := dashboard.NewGetMemberGrowthTrendLogic(r.Context(), svcCtx)
		resp, err := l.GetMemberGrowthTrend(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
