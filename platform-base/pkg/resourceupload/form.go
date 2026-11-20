// Package resourceupload 共用后台图片上传的表单、图片校验和 OSS 转发，不保存业务配置。
package resourceupload

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"oa.98ent.com/p9/common/xerr"
)

const (
	VIP                = "vipicon"              // VIP 图标场景，需提交等级下限。
	Logo               = "logo"                 // 分站 Logo 场景。
	Avatar             = "default_avatar"       // 总站默认头像场景。
	Frame              = "default_avatar_frame" // 总站默认头像框场景。
	AvatarCount        = 8                      // 初始化补齐八张头像，保留原有数量。
	FrameCount         = 6                      // 初始化补齐六张头像框，保留原有数量。
	ImageLimit   int64 = 200 << 10              // VIP、默认头像及头像框上限。
	LogoLimit    int64 = 32 << 20               // Logo 文件上限。
	FormOverhead int64 = 64 << 10               // 小图片表单的额外空间。
)

// Form 只保留允许的表单参数，不接受客户端指定站点、对象目录或对象文件名。
type Form struct {
	Scene    string                // 必填场景，由各 API 限制允许范围。
	MinLevel *int64                // 仅 VIP 必填，0 为有效值，不能以零值判断漏传。
	File     *multipart.FileHeader // 唯一文件，调用方负责清理 multipart 临时文件。
}

// ReadForm 限制请求总量，仅解析一个文件及场景参数，不解码图片或上传对象。
func ReadForm(w http.ResponseWriter, r *http.Request, maxBytes int64) (*Form, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		return nil, xerr.BadRequest("common.validation_error")
	}
	form := r.MultipartForm
	if form == nil || len(form.File) != 1 || len(form.File["file"]) != 1 || len(form.Value["scene"]) != 1 {
		return nil, xerr.BadRequest("common.validation_error")
	}
	result := &Form{
		Scene: form.Value["scene"][0], // 场景必须精确匹配服务端白名单。
		File:  form.File["file"][0],   // 只接收一个名为 file 的文件。
	}
	// VIP 仅接受 scene、min_level；其他场景仅接受 scene，拒绝额外和重复参数。
	if result.Scene != VIP {
		if len(form.Value) != 1 {
			return nil, xerr.BadRequest("common.validation_error")
		}
		return result, nil
	}
	if len(form.Value) != 2 || len(form.Value["min_level"]) != 1 {
		return nil, xerr.BadRequest("common.validation_error")
	}
	level, err := strconv.ParseInt(form.Value["min_level"][0], 10, 64)
	if err != nil || !validMinLevel(&level) {
		return nil, xerr.BadRequest("common.validation_error")
	}
	result.MinLevel = &level
	return result, nil
}

// validMinLevel 统一表单解析和上传前校验，VIP 等级下限仅允许 0、10…90。
func validMinLevel(level *int64) bool {
	return level != nil && *level >= 0 && *level <= 90 && *level%10 == 0
}
