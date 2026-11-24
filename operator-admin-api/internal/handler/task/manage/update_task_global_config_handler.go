// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/task/manage"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

// 更新任务全局配置（如生效时间>当前时间则延迟生效，否则立即生效）
func UpdateTaskGlobalConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateTaskGlobalConfigReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewUpdateTaskGlobalConfigLogic(r.Context(), svcCtx)
		resp, err := l.UpdateTaskGlobalConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
