package operator

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/common/response"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-base/api/internal/logic/operator"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
	"oa.98ent.com/p9/operator-base/api/internal/types"
)

// UpdateSiteLanguagesHandler 读取完整语言勾选列表，由 Logic 按登录站点保存。
func UpdateSiteLanguagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateSiteLanguagesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("invalid language selection request"))
			return
		}
		// 必须显式提交数组；[] 表示全部关闭，未传或 null 不能误当作关闭全部。
		if req.LanguageCodes == nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("language_codes must be an array"))
			return
		}
		l := operator.NewUpdateSiteLanguagesLogic(r.Context(), svcCtx)
		result, err := l.UpdateSiteLanguages(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		response.OkCtx(r.Context(), w, result)
	}
}
