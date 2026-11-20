// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vipdefaulticon

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/platform-base/api/internal/logic/vipdefaulticon"
	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
)

// ListAllVipDefaultIconsHandler 查询固定十组默认 VIP 图标，返回可展示的公开地址。
func ListAllVipDefaultIconsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 沿用统一请求解析入口，本接口无需分页或等级筛选参数。
		var req types.ListAllVipDefaultIconsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 创建当前请求的 Logic，再由查询方法调用默认 VIP 图标 RPC。
		l := vipdefaulticon.NewListAllVipDefaultIconsLogic(r.Context(), svcCtx)
		resp, err := l.ListAllVipDefaultIcons(&req)
		// 查询失败交给统一错误响应，成功返回图标分组列表。
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
