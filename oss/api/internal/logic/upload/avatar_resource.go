package upload

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/oss/common/object"
)

// isAvatarResource 两种总站公开素材使用独立场景，不影响原有 public、logo、vipicon。
func isAvatarResource(scene string) bool {
	return scene == "default_avatar" || scene == "default_avatar_frame"
}

// planAvatarResourceUpload 按现有素材校验 PNG、尺寸、大小和头像框透明效果。
func planAvatarResourceUpload(ctx context.Context, scene, extension string, content []byte) (string, error) {
	// OSS 根据透传的认证上下文再次确认总站身份，不只信任上游接口的校验。
	claims := ctxdata.ClaimsFromCtx(ctx)
	if claims == nil || claims.UserID <= 0 || !claims.IsPlatform || claims.TokenType == ctxdata.TokenTypePreview {
		return "", xerr.Forbidden("only platform users may upload default avatar resources")
	}
	// 只接受头像或头像框场景的 PNG 文件，单张最大 200KB。
	if !isAvatarResource(scene) || !strings.EqualFold(extension, ".png") || len(content) == 0 || len(content) > 200<<10 {
		return "", xerr.BadRequest("avatar resources must be PNG images up to 200KB")
	}
	if err := validateAvatarImage(scene, content); err != nil {
		return "", err
	}
	// 复用 OSS 对象路径规划，头像与头像框目录保持独立，不写入素材配置表。
	return object.PlanAvatarResourceKey(scene)
}

// validateAvatarImage 先限制解码尺寸，再完整读取 PNG；不裁剪、缩放或改变上传原图。
func validateAvatarImage(scene string, content []byte) error {
	size := 200
	if scene == "default_avatar_frame" {
		size = 145
	}
	// 根据场景检查固定尺寸，先检查尺寸再完整解码，限制解码内存。
	config, err := png.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width != size || config.Height != size {
		return xerr.BadRequest("avatar dimensions must be 200 by 200 pixels; avatar frames must be 145 by 145 pixels")
	}
	// 验证整个 PNG 可解码，头像框另需中心像素透明且存在可见图案。
	picture, err := png.Decode(bytes.NewReader(content))
	if err != nil {
		return xerr.BadRequest("avatar resource must be a complete PNG image")
	}
	if scene == "default_avatar_frame" {
		if !validAvatarFrame(picture, size) {
			return xerr.BadRequest("avatar frame must have a transparent center and a visible pattern")
		}
	}
	return nil
}

// validAvatarFrame 沿用中心像素透明且含可见图案的检查，不代表整个中间区域均透明。
func validAvatarFrame(picture image.Image, size int) bool {
	_, _, _, alpha := picture.At(size/2, size/2).RGBA()
	if alpha != 0 {
		return false
	}
	// 排除完全透明的空白图片，不要求头像框整个中间区域均透明。
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
