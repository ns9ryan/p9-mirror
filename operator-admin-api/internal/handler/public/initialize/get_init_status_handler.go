// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package initialize

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/public/initialize"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
)

// 检测分站代理线与等级是否完成初始化
func GetInitStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := initialize.NewGetInitStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetInitStatus()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
