package upload

import (
	"net/http"
	"strings"

	ulogic "oa.98ent.com/p9/oss/api/internal/logic/upload"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ResolveHandler 用 key（或已保存的 url）换当前可打开地址。
// @Tags 上传
// @Summary 按 key 取访问地址
// @Produce json
// @Param Authorization header string false "Bearer access_token"
// @Param key query string false "对象 key，推荐"
// @Param url query string false "已保存的访问 url；与 key 同时传以 key 为准"
// @Success 200 {object} types.FileReply
// @Router /url/detail [get]
func ResolveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResolveReq
		_ = httpx.Parse(r, &req)
		raw := strings.TrimSpace(req.Key)
		if raw == "" {
			raw = strings.TrimSpace(req.URL)
		}
		l := ulogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Resolve(raw)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
