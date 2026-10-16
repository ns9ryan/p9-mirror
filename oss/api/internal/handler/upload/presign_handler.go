package upload

import (
	"net/http"

	ulogic "oa.98ent.com/p9/oss/api/internal/logic/upload"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"

	"oa.98ent.com/p9/common/xerr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// PresignHandler 签发大文件直传用的预签名 PUT。
// @Tags 预上传
// @Summary 1. 签发 PUT URL
// @Accept json
// @Produce json
// @Param body body types.PresignRequest true "filename、size、可选 scene"
// @Success 200 {object} types.PresignReply
// @Router /upload/presign [post]
func PresignHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PresignRequest
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("invalid request"))
			return
		}
		l := ulogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Presign(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// CompleteHandler 确认预签名直传完成。
// @Tags 预上传
// @Summary 3. 确认完成
// @Accept json
// @Produce json
// @Param body body types.CompleteRequest true "presign 返回的 key；scene 可选"
// @Success 200 {object} types.FileReply
// @Router /upload/complete [post]
func CompleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CompleteRequest
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("key is required"))
			return
		}
		l := ulogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Complete(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
