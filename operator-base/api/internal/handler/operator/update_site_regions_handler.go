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

// UpdateSiteRegionsHandler 读取完整地区勾选列表，由 Logic 按登录站点保存。
func UpdateSiteRegionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析前端提交的完整地区勾选列表，具体编码与分配范围由 Logic、RPC 校验。
		var req types.UpdateSiteRegionsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("invalid region selection request"))
			return
		}
		// 必须显式提交数组；[] 表示全部关闭，未传或 null 不能误当作关闭全部。
		if req.RegionCodes == nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("region_codes must be an array"))
			return
		}
		l := operator.NewUpdateSiteRegionsLogic(r.Context(), svcCtx)
		result, err := l.UpdateSiteRegions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// Logic 调用本站 RPC 保存成功后，复用公共成功响应返回完整勾选结果。
		response.OkCtx(r.Context(), w, result)
	}
}
