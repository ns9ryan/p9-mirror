// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package operator_game_allocation

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/platform-operator/api/internal/logic/operator_game_allocation"
	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
)

// PublishOperatorGameAllocation 发布游戏资源分配到指定分站
func PublishOperatorGameAllocationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishOperatorGameAllocationRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := operator_game_allocation.NewPublishOperatorGameAllocationLogic(r.Context(), svcCtx)
		resp, err := l.PublishOperatorGameAllocation(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
