package defaultavatarframe

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarframepb"
)

type UpdateDefaultAvatarFrameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateDefaultAvatarFrameLogic 绑定修改请求的上下文和日志，此时不修改素材。
func NewUpdateDefaultAvatarFrameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDefaultAvatarFrameLogic {
	return &UpdateDefaultAvatarFrameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateDefaultAvatarFrame 只修改已有默认头像框，修改人由 RPC 从认证上下文获取。
func (l *UpdateDefaultAvatarFrameLogic) UpdateDefaultAvatarFrame(
	req *types.UpdateDefaultAvatarFrameRequest,
) (*types.UpdateDefaultAvatarFrameResponse, error) {
	// 直接调用默认素材修改 RPC，图片目录、Key 转换和修改人由 RPC 校验处理。
	_, err := l.svcCtx.DefaultAvatarFrameRpc.Update(l.ctx, &defaultavatarframepb.UpdateDefaultAvatarFrameRequest{
		Id:       req.Id,       // 要修改的已有头像框主键，不创建新记录。
		ImageUrl: req.ImageUrl, // 本类图片的公开 URL 或 Key，由 RPC 校验并转换入库值。
	})
	if err != nil {
		return nil, err
	}
	// RPC 修改成功后返回空业务对象，不在 API 另行写库。
	return &types.UpdateDefaultAvatarFrameResponse{}, nil
}
