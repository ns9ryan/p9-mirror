package defaultavatarservicelogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/platform-base/pkg/i18nkey"
	"oa.98ent.com/p9/platform-base/pkg/resourceupload"
	"oa.98ent.com/p9/platform-base/pkg/resourceurl"
	"oa.98ent.com/p9/platform-base/pkg/rpc/grpcerror"
	"oa.98ent.com/p9/platform-base/rpc/internal/enterror"
	"oa.98ent.com/p9/platform-base/rpc/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb"
)

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update 只按已有主键替换图片，记录修改人，不新增记录或删除旧 OSS 文件。
func (l *UpdateLogic) Update(in *defaultavatarpb.UpdateDefaultAvatarRequest) (*defaultavatarpb.UpdateDefaultAvatarResponse, error) {
	// 从认证上下文校验总站后台身份，分站用户和预览令牌不能修改总站素材。
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	if claims == nil || claims.UserID <= 0 || !claims.IsPlatform || claims.TokenType == ctxdata.TokenTypePreview {
		return nil, status.Error(codes.PermissionDenied, i18nkey.Forbidden)
	}
	// 修改必须指定已有记录的正整数主键。
	if in == nil || in.Id <= 0 {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	// 复用素材校验，拒绝空图片、跨素材目录和临时签名 URL。
	value := strings.TrimSpace(in.ImageUrl)
	if !resourceupload.ValidAvatarValue(resourceupload.Avatar, value) {
		return nil, grpcerror.InvalidArgument(i18nkey.ValidationError)
	}
	// 复用公开素材转换：已知 OSS 前缀裁成 Key，查询时再拼接；旧外链按原地址兼容。
	stored, _, err := resourceurl.Value(value, l.svcCtx.Config.ResourceBaseURL)
	if err != nil {
		return nil, err
	}
	// UpdateOneID 在记录不存在时返回错误，不能通过修改接口创建第 9 张图片。
	err = l.svcCtx.DB.DefaultAvatar.
		UpdateOneID(in.Id).          // 仅定位已有头像，主键不存在时返回 NotFound。
		SetImageURL(stored).         // 保存转换后的 Key 或兼容的旧公开 URL。
		SetUpdatedBy(claims.UserID). // 修改人取自认证上下文，修改时间沿用 Schema 自动更新。
		Exec(l.ctx)                  // 执行单条更新，保留旧 OSS 图片。
	if err != nil {
		// 复用 Ent 错误转换，记录不存在时保留 NotFound 语义。
		return nil, enterror.Handle(l.Logger, err)
	}
	l.svcCtx.BaseData.DefaultAvatar.Invalidate(l.ctx)

	// 修改成功只返回空业务对象，不新增素材或删除旧 OSS 图片。
	return &defaultavatarpb.UpdateDefaultAvatarResponse{}, nil
}
