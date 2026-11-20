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

// ListAllDefaultAvatarsHandler 查询全部默认头像，返回可展示的公开地址。
func ListAllDefaultAvatarsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 按项目统一入口解析请求；本接口无需分页或业务筛选参数。
		var req types.ListAllDefaultAvatarsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest(i18nkey.InvalidRequest))
			return
		}
		// 创建当前请求的 Logic，再由 Logic 调用默认素材查询 RPC。
		l := defaultavatar.NewListAllDefaultAvatarsLogic(r.Context(), svcCtx)
		resp, err := l.ListAllDefaultAvatars(&req)
		// 查询失败交给统一错误响应，成功返回素材列表。
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
