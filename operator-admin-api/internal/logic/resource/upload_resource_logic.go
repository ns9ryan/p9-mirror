package resource

import (
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
	"oa.98ent.com/p9/platform-base/pkg/resourceupload"
)

// UploadResourceLogic 分站统一上传 VIP 图标及本站 Logo，不保存业务配置。
type UploadResourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadResourceLogic 绑定认证上下文、请求日志及现有分站客户端。
func NewUploadResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadResourceLogic {
	return &UploadResourceLogic{
		Logger: logx.WithContext(ctx), // 日志关联当前上传请求。
		ctx:    ctx,                   // 透传登录身份、令牌及请求期限。
		svcCtx: svcCtx,                // 复用本站 OperatorBase RPC 和 OSS 配置。
	}
}

// UploadResource 只取 Token 中的站点，拒绝总站、代理、预览身份和任意路径参数。
func (l *UploadResourceLogic) UploadResource(form *resourceupload.Form) (*types.UploadResourceResponse, error) {
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.Forbidden("common.forbidden")
	}
	if claims.IsPlatform || claims.TokenType == ctxdata.TokenTypePreview ||
		!resourceupload.ValidOperatorCode(claims.OperatorCode) || ctxdata.RawTokenFromCtx(l.ctx) == "" {
		return nil, xerr.Forbidden("common.forbidden")
	}
	if form == nil || (form.Scene != resourceupload.VIP && form.Scene != resourceupload.Logo) {
		return nil, xerr.BadRequest("common.validation_error")
	}
	// Logo 先调用本站 RPC 核对登录站点，再打开、解码及上传文件；VIP 不增加这项依赖。
	if form.Scene == resourceupload.Logo {
		if err := l.checkSite(claims.OperatorCode); err != nil {
			return nil, err
		}
	}
	// 复用统一图片校验和 HTTP 转发；目标 OSS 必须连接同一分站 Core 认证令牌。
	result, err := resourceupload.Upload(l.ctx, l.svcCtx.Config.OssApiBaseURL, form, claims.OperatorCode, l.Logger)
	if err != nil {
		return nil, err
	}
	return &types.UploadResourceResponse{
		Url:          result.Url,          // 已核对本站归属的永久地址。
		Key:          result.Key,          // 原场景及本站目录的对象 Key。
		Scene:        result.Scene,        // 本次选定的场景。
		Visibility:   result.Visibility,   // 固定公开。
		UrlExpiresIn: result.UrlExpiresIn, // 永久地址为 0。
	}, nil
}

// checkSite 只在 Logo 请求内调用本站 OperatorBase RPC，不改变服务启动条件或写入站点配置。
func (l *UploadResourceLogic) checkSite(operatorCode string) error {
	if strings.TrimSpace(l.svcCtx.Config.OssApiBaseURL) == "" || l.svcCtx.OperatorBaseRpc == nil {
		return xerr.Err(http.StatusServiceUnavailable, "common.service_unavailable")
	}
	result, err := l.svcCtx.OperatorBaseRpc.GetSiteConfig(l.ctx, &operatorpb.GetSiteConfigRequest{})
	if err != nil {
		l.Errorw("Logo 上传核对本站配置失败", logx.Field("operator_code", operatorCode), logx.Field("error", err.Error()))
		// 保留下游参数、权限和服务不可用语义，未知内部信息不直接返回前端。
		switch status.Code(err) {
		case codes.InvalidArgument:
			return xerr.BadRequest("common.validation_error")
		case codes.Unauthenticated:
			return xerr.Err(http.StatusUnauthorized, "common.unauthorized")
		case codes.PermissionDenied:
			return xerr.Forbidden("common.forbidden")
		case codes.NotFound:
			return xerr.Err(http.StatusNotFound, "common.data_not_found")
		case codes.Unavailable:
			return xerr.Err(http.StatusServiceUnavailable, "common.service_unavailable")
		case codes.DeadlineExceeded:
			return xerr.Err(http.StatusGatewayTimeout, "common.service_unavailable")
		default:
			return xerr.InternalServerError("common.internal_error")
		}
	}
	if result.GetSiteConfig().GetOperatorCode() != operatorCode {
		return xerr.Forbidden("common.forbidden")
	}
	return nil
}
