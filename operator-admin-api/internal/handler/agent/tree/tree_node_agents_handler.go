// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tree

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/agent/tree"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

// 代理树节点（按上级 ID；0 为顶层）
func TreeNodeAgentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TreeNodeAgentsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tree.NewTreeNodeAgentsLogic(r.Context(), svcCtx)
		resp, err := l.TreeNodeAgents(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
