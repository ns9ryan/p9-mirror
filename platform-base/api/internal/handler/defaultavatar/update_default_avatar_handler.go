package defaultavatar

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-base/api/internal/logic/defaultavatar"
	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
)

// UpdateDefaultAvatarHandler 解析已有默认头像的修改请求，不接收新增或删除操作。
func UpdateDefaultAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 JSON 请求并执行类型上的参数校验，读取记录 ID 和图片 URL 或 Key。
		var req types.UpdateDefaultAvatarRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest(i18nkey.InvalidRequest))
			return
		}
		// 创建修改 Logic；实际 RPC 在下方方法中发起，身份与图片归属由 RPC 校验。
		l := defaultavatar.NewUpdateDefaultAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UpdateDefaultAvatar(&req)
		// 失败交给统一错误响应；修改成功返回空业务对象。
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
