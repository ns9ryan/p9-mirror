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

type SaveVipDefaultIconsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSaveVipDefaultIconsLogic 绑定图标保存的上下文和日志，此时不写入配置。
func NewSaveVipDefaultIconsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveVipDefaultIconsLogic {
	return &SaveVipDefaultIconsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SaveVipDefaultIcons 保存本次提交的默认图标组。
func (l *SaveVipDefaultIconsLogic) SaveVipDefaultIcons(
	req *types.SaveVipDefaultIconsRequest,
) (*types.SaveVipDefaultIconsResponse, error) {
	// 只转换本次提交的等级组；未提交的组不补入请求，由 RPC 保持原值。
	list := make([]*vipdefaulticonpb.VipDefaultIconGroupInfo, 0, len(req.List))
	for _, item := range req.List {
		list = append(list, &vipdefaulticonpb.VipDefaultIconGroupInfo{
			MinLevel: item.MinLevel, // 本次保存组的等级下限，0 级有效。
			MaxLevel: item.MaxLevel, // 等级上限，由 RPC 核对是否等于下限加 9。
			IconUrl:  item.IconUrl,  // 接受本组 Key 或公开 URL，空串表示清空。
		})
	}
	// 直接调用默认图标保存 RPC，地址转 Key；修改人随认证上下文透传，客户端不能自行指定。
	_, err := l.svcCtx.VipDefaultIconRpc.Save(l.ctx, &vipdefaulticonpb.SaveVipDefaultIconsRequest{
		List: list, // 只提交本次修改的组，未提交组保持原值。
	})
	if err != nil {
		return nil, err
	}
	// RPC 事务提交成功后返回空业务对象，失败时保留原错误。
	return &types.SaveVipDefaultIconsResponse{}, nil
}
