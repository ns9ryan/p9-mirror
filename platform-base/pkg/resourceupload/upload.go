package resourceupload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
)

var (
	operatorPattern     = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)
	iconFilenamePattern = regexp.MustCompile(`^[0-9a-f]{32}\.png$`)
	avatarKey           = regexp.MustCompile(`^avatar-resource/avatars/(?:[0-9a-f]{32}|avatar[1-8]-[0-9a-f]{16})\.png$`)
	frameKey            = regexp.MustCompile(
		`^avatar-resource/avatar-frames/(?:[0-9a-f]{32}|frames[1-6]-[0-9a-f]{16})\.png$`,
	)
)

// Result 是经场景、归属和永久公开地址核对后的素材；不代表配置已保存。
type Result struct {
	Url          string `json:"url"`            // 完整公开地址，可预览或提交保存。
	Key          string `json:"key"`            // 桶内对象路径，不含域名及桶前缀。
	Scene        string `json:"scene"`          // 实际上传场景。
	Visibility   string `json:"visibility"`     // 只允许 public。
	UrlExpiresIn int64  `json:"url_expires_in"` // 固定为 0，表示永久公开地址。
}

// ValidOperatorCode 限制站点目录片段，不允许路径分隔符及相对路径。
func ValidOperatorCode(code string) bool {
	return operatorPattern.MatchString(code)
}

// Upload 在 API 完成身份和场景授权后校验图片并调用 OSS；无重试，不保存配置或删除旧图片。
// operatorCode 为空表示总站素材，否则只接受当前分站目录；Logger 复用调用方请求日志。
func Upload(ctx context.Context, baseURL string, form *Form, operatorCode string, logger logx.Logger) (*Result, error) {
	content, err := readImage(form)
	if err != nil {
		return nil, err
	}
	request, err := uploadRequest(ctx, baseURL, form, content)
	if err != nil {
		return nil, err
	}
	// Logo 保留 60 秒超时，其余素材 30 秒；禁止跟随重定向及自动重传对象。
	timeout := 30 * time.Second
	if form.Scene == Logo {
		timeout = 60 * time.Second
	}
	client := &http.Client{
		Timeout: timeout, // 仍受上层请求取消和超时约束。
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Do(request)
	if err != nil {
		logger.Errorw("后台图片调用 OSS 失败", logx.Field("scene", form.Scene), logx.Field("error", err.Error()))
		return nil, xerr.Err(http.StatusBadGateway, "resource.uploadFailed")
	}
	defer response.Body.Close()
	result, err := decodeReply(response)
	if err != nil {
		logger.Errorw("后台图片 OSS 响应失败", logx.Field("scene", form.Scene),
			logx.Field("http_status", response.StatusCode))
		return nil, err
	}
	if !validResult(result, form, operatorCode) {
		logger.Errorw("后台图片 OSS 返回素材不匹配", logx.Field("scene", form.Scene),
			logx.Field("operator_code", operatorCode))
		return nil, xerr.Err(http.StatusBadGateway, "resource.uploadFailed")
	}
	return result, nil
}

// readImage 在站点校验之后打开文件，按场景限制大小并完整解码，不信任扩展名或表单大小。
func readImage(form *Form) ([]byte, error) {
	if form == nil || form.File == nil {
		return nil, xerr.BadRequest("common.validation_error")
	}
	limit := ImageLimit
	if form.Scene == Logo {
		limit = LogoLimit
	}
	if form.File.Size <= 0 || form.File.Size > limit {
		return nil, xerr.BadRequest("resource.invalidFile")
	}
	file, err := form.File.Open()
	if err != nil {
		return nil, xerr.BadRequest("resource.invalidFile")
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil || len(content) == 0 || int64(len(content)) > limit {
		return nil, xerr.BadRequest("resource.invalidFile")
	}
	if !validImage(form, content) {
		return nil, xerr.BadRequest("resource.invalidFile")
	}
	return content, nil
}

// validImage 保留各类原始规格，不缩放或转换图片；未知场景不能上传。
func validImage(form *Form, content []byte) bool {
	switch form.Scene {
	case Avatar, Frame:
		return validAvatarFile(form.Scene, form.File.Filename, content)
	case VIP:
		if !validMinLevel(form.MinLevel) {
			return false
		}
		if !strings.EqualFold(filepath.Ext(form.File.Filename), ".png") {
			return false
		}
		config, err := png.DecodeConfig(bytes.NewReader(content))
		if err != nil || config.Width != 115 || config.Height != 165 {
			return false
		}
		_, err = png.Decode(bytes.NewReader(content))
		return err == nil
	case Logo:
		return validLogo(form.File.Filename, content)
	default:
		return false
	}
}

// validLogo 保留 JPG/PNG、每边不超过 4096 像素及完整解码的原有约束。
func validLogo(filename string, content []byte) bool {
	formats := map[string]string{
		".jpg": "jpeg", // 保持原 Logo 接口支持的扩展名。
		".png": "png",  // 扩展名还须与实际格式一致。
	}
	expected := formats[strings.ToLower(filepath.Ext(filename))]
	config, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || expected == "" || format != expected {
		return false
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > 4096 || config.Height > 4096 {
		return false
	}
	_, decodedFormat, err := image.Decode(bytes.NewReader(content))
	return err == nil && decodedFormat == format
}

// uploadRequest 仅使用服务端配置的 OSS 地址，透传认证链确认的令牌、IP 和语言。
func uploadRequest(ctx context.Context, baseURL string, form *Form, content []byte) (*http.Request, error) {
	endpoint, err := resourceurl.PublicURL(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "common.service_unavailable")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/p9/oss/file/upload"
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(form.File.Filename))
	if err != nil {
		return nil, xerr.InternalServerError("common.internal_error")
	}
	if _, err := part.Write(content); err != nil {
		return nil, xerr.InternalServerError("common.internal_error")
	}
	if err := writer.WriteField("scene", form.Scene); err != nil {
		return nil, xerr.InternalServerError("common.internal_error")
	}
	if form.Scene == VIP {
		if err := writer.WriteField("min_level", strconv.FormatInt(*form.MinLevel, 10)); err != nil {
			return nil, xerr.InternalServerError("common.internal_error")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, xerr.InternalServerError("common.internal_error")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), &body)
	if err != nil {
		return nil, xerr.InternalServerError("common.internal_error")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+ctxdata.RawTokenFromCtx(ctx))
	if ip := ctxdata.ClientIPFromCtx(ctx); ip != "" {
		request.Header.Set("X-Forwarded-For", ip)
	}
	request.Header.Set("X-Lang", i18n.Lang(ctx, nil))
	return request, nil
}

// decodeReply 同时核对 HTTP 状态、显式成功业务码和响应上限；隐藏下游内部错误。
func decodeReply(response *http.Response) (*Result, error) {
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		code := response.StatusCode
		if code < 400 || code >= 500 {
			code = http.StatusBadGateway
		}
		return nil, xerr.Err(code, "resource.uploadFailed")
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(content) > 1<<20 {
		return nil, xerr.Err(http.StatusBadGateway, "resource.uploadFailed")
	}
	var reply struct {
		Code *int    `json:"code"` // 缺失业务码不能视为成功。
		Data *Result `json:"data"` // 缺失素材或归属不符均拒绝返回。
	}
	if err := json.Unmarshal(content, &reply); err != nil {
		return nil, xerr.Err(http.StatusBadGateway, "resource.uploadFailed")
	}
	if reply.Code == nil || *reply.Code != 0 || reply.Data == nil {
		return nil, xerr.Err(http.StatusBadGateway, "resource.uploadFailed")
	}
	return reply.Data, nil
}

// validResult 只接受同类、同站点、同等级目录的永久公开素材，URL 路径须与 Key 对应。
func validResult(result *Result, form *Form, operatorCode string) bool {
	locator, err := resourceurl.PublicURL(result.Url)
	if err != nil || len(result.Url) > 1024 || result.Scene != form.Scene ||
		result.Visibility != "public" || result.UrlExpiresIn != 0 {
		return false
	}
	if !strings.HasSuffix(locator.Path, "/"+result.Key) {
		return false
	}
	return validKey(result.Key, form, operatorCode)
}

// validKey 保留各场景的目录规则，VIP 等级及分站编码不得跨组或跨站使用。
func validKey(key string, form *Form, operatorCode string) bool {
	switch form.Scene {
	case Avatar, Frame:
		return operatorCode == "" && validAvatarKey(form.Scene, key)
	case VIP:
		prefix := fmt.Sprintf("vipicon/%d-%d/", *form.MinLevel, *form.MinLevel+9)
		if operatorCode != "" {
			prefix = fmt.Sprintf("vipicon/%s/%d-%d/", operatorCode, *form.MinLevel, *form.MinLevel+9)
		}
		return strings.HasPrefix(key, prefix) &&
			iconFilenamePattern.MatchString(strings.TrimPrefix(key, prefix))
	case Logo:
		// 兼容原 Logo 两种目录；完整 URL 检查同时拒绝 Key 中的相对路径和转义。
		return ValidOperatorCode(operatorCode) &&
			(strings.HasPrefix(key, "logo/"+operatorCode+"/") ||
				strings.HasPrefix(key, operatorCode+"/logo/"))
	default:
		return false
	}
}

// validAvatarKey 同时兼容首次导入的原图名称和后续上传生成的随机文件名。
func validAvatarKey(scene, key string) bool {
	switch scene {
	case Avatar:
		return avatarKey.MatchString(key)
	case Frame:
		return frameKey.MatchString(key)
	default:
		return false
	}
}

// validAvatarURL 只接收对应目录的 HTTP(S) 永久地址，禁止清空或使用临时签名链接。
func validAvatarURL(scene, value string) bool {
	parsed, err := resourceurl.PublicURL(value)
	if err != nil || len(value) > 1024 {
		return false
	}
	// 提取素材目录后的对象 Key，再按头像或头像框的目录及文件名规则检查。
	position := strings.LastIndex(parsed.Path, "/avatar-resource/")
	return position >= 0 && validAvatarKey(scene, parsed.Path[position+1:])
}

// ValidAvatarValue 供头像及头像框保存接口复用，同时接受 Key 和永久 URL，拒绝清空、跨类和临时地址。
func ValidAvatarValue(scene, value string) bool {
	return validAvatarKey(scene, value) || validAvatarURL(scene, value)
}

// validAvatarFile 先限制大小和像素再完整解码，拒绝伪造、截断文件及遮挡头像的实心头像框。
func validAvatarFile(scene, filename string, content []byte) bool {
	// 限制文件大小并检查 PNG 扩展名，实际图片内容还需继续解码验证。
	if len(content) == 0 || int64(len(content)) > ImageLimit || !strings.EqualFold(filepath.Ext(filename), ".png") {
		return false
	}
	// 头像固定为 200×200，头像框固定为 145×145，不在上传时自动缩放。
	size := 200
	switch scene {
	case Avatar:
	case Frame:
		size = 145
	default:
		return false
	}
	// 先读取尺寸并限制解码范围，避免仅凭扩展名放行伪造或超大图片。
	config, err := png.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width != size || config.Height != size {
		return false
	}
	// 完整解码确认 PNG 未截断或损坏，头像框还要继续检查透明效果。
	picture, err := png.Decode(bytes.NewReader(content))
	if err != nil {
		return false
	}
	if scene == Frame {
		return validFrame(picture, size)
	}
	return true
}

// validFrame 沿用中心点透明且存在可见图案的规则，不把头像和头像框合成图片。
func validFrame(picture image.Image, size int) bool {
	// 只检查中心像素完全透明，不代表整个中央区域均透明。
	_, _, _, alpha := picture.At(size/2, size/2).RGBA()
	if alpha != 0 {
		return false
	}
	// 至少存在一个非全透明像素，排除整张透明、没有可见图案的头像框。
	for row := 0; row < size; row++ {
		for column := 0; column < size; column++ {
			_, _, _, alpha := picture.At(column, row).RGBA()
			if alpha != 0 {
				return true
			}
		}
	}
	return false
}
