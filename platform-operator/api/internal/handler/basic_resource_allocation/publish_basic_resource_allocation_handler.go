// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package basic_resource_allocation

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/platform-operator/api/internal/logic/basic_resource_allocation"
	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
)

// PublishBasicResourceAllocation 发布基础资源分配到指定分站
func PublishBasicResourceAllocationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishBasicResourceAllocationRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := basic_resource_allocation.NewPublishBasicResourceAllocationLogic(r.Context(), svcCtx)
		resp, err := l.PublishBasicResourceAllocation(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
