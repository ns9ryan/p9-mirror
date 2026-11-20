package defaultavatar

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/defaultavatarpb"
)

type ListAllDefaultAvatarsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListAllDefaultAvatarsLogic 绑定查询请求的上下文和日志，此时不调用 RPC。
func NewListAllDefaultAvatarsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAllDefaultAvatarsLogic {
	return &ListAllDefaultAvatarsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListAllDefaultAvatars 读取默认头像列表，返回数据库中的主键和公开永久地址。
func (l *ListAllDefaultAvatarsLogic) ListAllDefaultAvatars(
	req *types.ListAllDefaultAvatarsRequest,
) (*types.ListAllDefaultAvatarsResponse, error) {
	// 直接调用总站默认素材 RPC，公开地址已由 RPC 根据保存的 Key 和配置拼接。
	result, err := l.svcCtx.DefaultAvatarRpc.ListAll(l.ctx, &defaultavatarpb.ListAllDefaultAvatarsRequest{})
	if err != nil {
		return nil, err
	}
	// 将 RPC 返回值逐项映射为 API 响应；空列表保持为 []，图片地址不在 API 重复拼接。
	list := make([]types.DefaultAvatarInfo, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.DefaultAvatarInfo{
			Id:       item.Id,       // 已有头像记录的主键，修改时用于定位。
			ImageUrl: item.ImageUrl, // RPC 已处理的公开地址，空串表示未配置。
		})
	}
	// 返回前端列表，保留 RPC 已处理的图片地址和默认值。
	return &types.ListAllDefaultAvatarsResponse{
		List: list, // 按主键升序返回头像列表，空列表保持为 []。
	}, nil
}
