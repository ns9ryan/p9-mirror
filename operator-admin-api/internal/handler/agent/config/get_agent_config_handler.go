// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"oa.98ent.com/p9/operator-admin-api/internal/logic/agent/config"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
)

// 获取代理配置
func GetAgentConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := config.NewGetAgentConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetAgentConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
