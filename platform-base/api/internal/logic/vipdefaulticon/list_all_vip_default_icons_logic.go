// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package vipdefaulticon

import (
	"context"

	"oa.98ent.com/p9/platform-base/api/internal/svc"
	"oa.98ent.com/p9/platform-base/api/internal/types"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/vipdefaulticonpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAllVipDefaultIconsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListAllVipDefaultIconsLogic 绑定图标查询的上下文和日志，此时不调用 RPC。
func NewListAllVipDefaultIconsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAllVipDefaultIconsLogic {
	return &ListAllVipDefaultIconsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListAllVipDefaultIcons 获取固定十组默认图标。
func (l *ListAllVipDefaultIconsLogic) ListAllVipDefaultIcons(
	req *types.ListAllVipDefaultIconsRequest,
) (*types.ListAllVipDefaultIconsResponse, error) {
	// 调用默认图标查询 RPC。
	result, err := l.svcCtx.VipDefaultIconRpc.ListAll(l.ctx, &vipdefaulticonpb.ListAllVipDefaultIconsRequest{})
	if err != nil {
		return nil, err
	}
	// 将 RPC 返回值逐项映射为 API 响应；空列表保持为 []，图片地址不在 API 重复拼接。
	list := make([]types.VipDefaultIconGroupInfo, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.VipDefaultIconGroupInfo{
			MinLevel: item.MinLevel, // 本组 VIP 等级下限。
			MaxLevel: item.MaxLevel, // 本组 VIP 等级上限，为下限加 9。
			IconUrl:  item.IconUrl,  // RPC 已处理的公开地址，空串表示未配置。
		})
	}
	// 返回前端列表，保留 RPC 已处理的图片地址和默认值。
	return &types.ListAllVipDefaultIconsResponse{
		List: list, // 按等级下限升序返回固定十组图标。
	}, nil
}
