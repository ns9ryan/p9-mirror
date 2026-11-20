package resource

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/platform-base/api/internal/logic/resource"
	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/pkg/resourceupload"
)

// UploadResourceHandler 解析统一表单，业务身份、图片及 OSS 返回值由 Logic 校验。
func UploadResourceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, err := resourceupload.ReadForm(w, r, resourceupload.ImageLimit+resourceupload.FormOverhead)
		// 无论校验或上传是否成功，都只清理本次 multipart 临时文件。
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := resource.NewUploadResourceLogic(r.Context(), svcCtx)
		resp, err := l.UploadResource(form)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
