package resource

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/pkg/resourceupload"
)

// UploadResourceLogic 总站素材上传，不保存配置或删除旧图片。
type UploadResourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadResourceLogic 绑定当前请求日志和服务端 OSS 配置。
func NewUploadResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadResourceLogic {
	return &UploadResourceLogic{
		Logger: logx.WithContext(ctx), // 关联本次请求日志。
		ctx:    ctx,                   // 携带认证身份及请求取消信号。
		svcCtx: svcCtx,                // 复用已配置的总站 OSS 地址。
	}
}

// UploadResource 仅允许总站后台上传三类公开素材，分站、代理和预览身份不能使用。
func (l *UploadResourceLogic) UploadResource(form *resourceupload.Form) (*types.UploadResourceResponse, error) {
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil || claims.UserID <= 0 || !claims.IsPlatform || claims.TokenType == ctxdata.TokenTypePreview {
		return nil, xerr.Forbidden("common.forbidden")
	}
	if ctxdata.RawTokenFromCtx(l.ctx) == "" {
		return nil, xerr.Forbidden("common.forbidden")
	}
	// 场景白名单不映射为任意目录，新增类型须在后端明确接入。
	if form == nil {
		return nil, xerr.BadRequest("common.validation_error")
	}
	switch form.Scene {
	case resourceupload.VIP, resourceupload.Avatar, resourceupload.Frame:
	default:
		return nil, xerr.BadRequest("common.validation_error")
	}
	// 复用统一图片检查及 HTTP 转发；OSS 仍验证总站 Token 和上传权限。
	result, err := resourceupload.Upload(l.ctx, l.svcCtx.Config.OssApiBaseURL, form, "", l.Logger)
	if err != nil {
		return nil, err
	}
	return &types.UploadResourceResponse{
		Url:          result.Url,          // 已核对的永久公开地址。
		Key:          result.Key,          // 原有场景对应的桶内路径。
		Scene:        result.Scene,        // 与本次请求一致的场景。
		Visibility:   result.Visibility,   // 固定公开。
		UrlExpiresIn: result.UrlExpiresIn, // 永久地址为 0。
	}, nil
}
