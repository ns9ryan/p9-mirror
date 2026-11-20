// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package phone

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/public/phone"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
)

// 手机区号列表
func ListPhoneCallingCodesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := phone.NewListPhoneCallingCodesLogic(r.Context(), svcCtx)
		resp, err := l.ListPhoneCallingCodes()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
