package upload

import (
	"bytes"
	"context"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/oss/common/object"
)

// FormMinLevel 只从 multipart 正文读取等级下限，保留 0 级与未传参数的区别。
func FormMinLevel(request *http.Request) (*int64, error) {
	if request.MultipartForm == nil {
		return nil, nil
	}
	// 只读取 multipart 正文，未传返回 nil，与明确提交 0 级区分开。
	values, exists := request.MultipartForm.Value["min_level"]
	if !exists {
		return nil, nil
	}
	if len(values) != 1 {
		return nil, xerr.BadRequest("min_level must be provided once")
	}
	minLevel, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return nil, xerr.BadRequest("invalid min_level")
	}
	return &minLevel, nil
}

// planVipIconUpload 校验固定等级组图标；总站默认素材和分站素材使用不同目录。
func planVipIconUpload(ctx context.Context, minLevel *int64, extension string, content []byte) (string, error) {
	// OSS 独立校验身份；分站只能写入令牌对应站点目录，预览令牌不能上传。
	claims := ctxdata.ClaimsFromCtx(ctx)
	if claims == nil || claims.UserID <= 0 || claims.TokenType == ctxdata.TokenTypePreview {
		return "", xerr.Forbidden("login user is required to upload VIP icons")
	}
	if !claims.IsPlatform {
		if !claims.UserType.IsOperator() || object.SanitizeOperatorCode(claims.OperatorCode) == "" {
			return "", xerr.Forbidden("login operator is required to upload VIP icons")
		}
	}
	// 校验固定等级组和文件基础限制，再检查真实 PNG 尺寸与完整性。
	if !validVipIconLevel(minLevel) {
		return "", xerr.BadRequest("min_level must be 0, 10, 20, 30, 40, 50, 60, 70, 80 or 90")
	}
	if !strings.EqualFold(extension, ".png") || len(content) == 0 || len(content) > 200<<10 {
		return "", xerr.BadRequest("VIP icon must be a PNG image up to 200KB")
	}
	if err := validateVipIconImage(content); err != nil {
		return "", err
	}
	// 复用 OSS 对象路径规划，固定等级范围与原有随机文件名规则保持一致。
	if !claims.IsPlatform {
		return object.PlanOperatorVipIconKey(claims.OperatorCode, *minLevel)
	}
	return object.PlanVipIconKey(*minLevel)
}

// validVipIconLevel 区分未传与有效的 0 级，只允许固定十组的等级下限。
func validVipIconLevel(minLevel *int64) bool {
	return minLevel != nil && *minLevel >= 0 && *minLevel <= 90 && *minLevel%10 == 0
}

// validateVipIconImage 校验真实 PNG 尺寸与完整性，不只依据扩展名和文件头放行。
func validateVipIconImage(content []byte) error {
	// 先限制解码尺寸，再完整解码，避免伪造文件头或损坏图片进入图标目录。
	config, err := png.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width != 115 || config.Height != 165 {
		return xerr.BadRequest("VIP icon dimensions must be 115 by 165 pixels")
	}
	if _, err := png.Decode(bytes.NewReader(content)); err != nil {
		return xerr.BadRequest("VIP icon must be a complete PNG image")
	}
	return nil
}
