package file

import (
	"net/http"

	flogic "oa.98ent.com/p9/oss/api/internal/logic/file"
	"oa.98ent.com/p9/oss/api/internal/svc"
	"oa.98ent.com/p9/oss/api/internal/types"

	"oa.98ent.com/p9/common/xerr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ListHandler 文件档案列表。
// @Tags 文件管理
// @Summary 文件列表
// @Accept json
// @Produce json
// @Param body body types.FileListReq false "scene、page、page_size"
// @Success 200 {object} types.FileListReply
// @Router /file/list [post]
func ListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileListReq
		_ = httpx.ParseJsonBody(r, &req)
		l := flogic.NewListLogic(r.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// DeleteHandler 删除对象并删档案行。
// @Tags 文件管理
// @Summary 删除文件
// @Accept json
// @Produce json
// @Param body body types.FileDeleteReq true "对象 key"
// @Success 200 {object} types.FileReply
// @Router /file/delete [post]
func DeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileDeleteReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.BadRequest("key is required"))
			return
		}
		l := flogic.NewDeleteLogic(r.Context(), svcCtx)
		resp, err := l.Delete(req.Key)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
