package operator

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/common/response"
	"oa.98ent.com/p9/operator-base/api/internal/logic/operator"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
)

// GetSiteConfigHandler 查询登录站点配置，站点身份由认证中间件写入上下文。
func GetSiteConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := operator.NewGetSiteConfigLogic(r.Context(), svcCtx)
		result, err := l.GetSiteConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		response.OkCtx(r.Context(), w, result)
	}
}
