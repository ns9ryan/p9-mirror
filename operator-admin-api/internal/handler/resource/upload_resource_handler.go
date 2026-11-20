package resource

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"oa.98ent.com/p9/operator-admin-api/internal/logic/resource"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/platform-base/pkg/resourceupload"
)

// UploadResourceHandler 解析唯一文件和场景，身份、站点及图片检查在 Logic 完成。
func UploadResourceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form, err := resourceupload.ReadForm(w, r, resourceupload.LogoLimit+(1<<20))
		// 大文件可能落临时磁盘；仅清理本次表单文件，不删除 OSS 对象。
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
