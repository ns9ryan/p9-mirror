// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package nickname

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-base/api/internal/logic/nickname"
	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
)

func ListNicknamesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListNicknamesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest(i18nkey.InvalidRequest))
			return
		}

		l := nickname.NewListNicknamesLogic(r.Context(), svcCtx)
		resp, err := l.ListNicknames(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
