package defaultavatar

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb"
)

type UpdateDefaultAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateDefaultAvatarLogic 绑定修改请求的上下文和日志，此时不修改素材。
func NewUpdateDefaultAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDefaultAvatarLogic {
	return &UpdateDefaultAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateDefaultAvatar 只修改已有默认头像，修改人由 RPC 从认证上下文获取。
func (l *UpdateDefaultAvatarLogic) UpdateDefaultAvatar(
	req *types.UpdateDefaultAvatarRequest,
) (*types.UpdateDefaultAvatarResponse, error) {
	// 直接调用默认素材修改 RPC，图片目录、Key 转换和修改人由 RPC 校验处理。
	_, err := l.svcCtx.DefaultAvatarRpc.Update(l.ctx, &defaultavatarpb.UpdateDefaultAvatarRequest{
		Id:       req.Id,       // 要修改的已有头像主键，不创建新记录。
		ImageUrl: req.ImageUrl, // 本类图片的公开 URL 或 Key，由 RPC 校验并转换入库值。
	})
	if err != nil {
		return nil, err
	}
	// RPC 修改成功后返回空业务对象，不在 API 另行写库。
	return &types.UpdateDefaultAvatarResponse{}, nil
}
