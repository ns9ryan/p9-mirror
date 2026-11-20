package operator

import (
	"net/http"
	"net/url"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/common/response"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-base/api/internal/logic/operator"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
	"oa.98ent.com/p9/operator-base/api/internal/types"
)

// GetSiteCodeHandler 公开读取 query 中的域名，不使用 Host 或登录身份确定查询目标。
func GetSiteCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().
			Set("Cache-Control", "no-store")
		// 查询目标必须明确提交一次，拒绝重复域名参数，避免调用方和服务端理解不一致。
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil || len(query["domain_name"]) != 1 {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("exactly one domain_name query parameter is required"))
			return
		}
		req := types.GetSiteCodeRequest{
			DomainName: query.Get("domain_name"),
		}
		// 构造 Logic 后再发起查询；站点编码由本站 RPC 按显式提交的域名查出。
		l := operator.NewGetSiteCodeLogic(r.Context(), svcCtx)
		result, err := l.GetSiteCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		response.OkCtx(r.Context(), w, result)
	}
}
