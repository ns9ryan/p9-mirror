package upload

import (
	"net/http"

	ulogic "oa.98ent.com/p9/oss/api/internal/logic/upload"
	"oa.98ent.com/p9/oss/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// UploadHandler 中转上传。
// @Tags 上传
// @Summary 上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param scene formData string false "场景枚举；不传则 default_scene。禁止传桶名"
// @Success 200 {object} types.FileReply
// @Router /file/upload [post]
func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		up := svcCtx.Upload
		name, content, err := ulogic.ReadMultipart(r, w, up.MaxSize, up.MultipartMemory)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := ulogic.NewLogic(r.Context(), svcCtx)
		resp, err := l.Upload(ulogic.FormScene(r), name, content)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
