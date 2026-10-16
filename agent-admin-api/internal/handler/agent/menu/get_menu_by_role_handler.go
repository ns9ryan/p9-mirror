// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package menu

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/agent-admin-api/internal/logic/agent/menu"
	"oa.98ent.com/p9/agent-admin-api/internal/svc"
)

// 当前菜单树
func GetMenuByRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := menu.NewGetMenuByRoleLogic(r.Context(), svcCtx)
		resp, err := l.GetMenuByRole()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
