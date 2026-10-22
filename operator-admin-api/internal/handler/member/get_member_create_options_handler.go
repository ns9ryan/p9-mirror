package member

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	memberlogic "oa.98ent.com/p9/operator-admin-api/internal/logic/member"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
)

func GetMemberCreateOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := memberlogic.NewGetMemberCreateOptionsLogic(r.Context(), svcCtx).GetMemberCreateOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
