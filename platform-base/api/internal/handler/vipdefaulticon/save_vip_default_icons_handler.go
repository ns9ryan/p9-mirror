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

// SaveVipDefaultIconsHandler 保存提交的默认 VIP 图标组，未提交组保持原值。
func SaveVipDefaultIconsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 JSON 并执行请求类型上的校验，具体等级组和图片规则由 RPC 继续校验。
		var req types.SaveVipDefaultIconsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 创建保存 Logic，再由保存方法调用 RPC；Handler 不直接写入数据库。
		l := vipdefaulticon.NewSaveVipDefaultIconsLogic(r.Context(), svcCtx)
		resp, err := l.SaveVipDefaultIcons(&req)
		// RPC 保存失败时返回错误，全部提交成功后返回空业务对象。
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
