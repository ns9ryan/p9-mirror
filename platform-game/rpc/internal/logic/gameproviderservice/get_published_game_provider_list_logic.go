package gameproviderservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublishedGameProviderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPublishedGameProviderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublishedGameProviderListLogic {
	return &GetPublishedGameProviderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取已发布的游戏供应商列表（供分站同步）
func (l *GetPublishedGameProviderListLogic) GetPublishedGameProviderList(in *platform_game.GetPublishedGameProviderListRequest) (*platform_game.GetPublishedGameProviderListResp, error) {
	// 验证请求参数
	if in.OpCode == "" {
		return nil, fmt.Errorf("op_code is required")
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	offset := (in.Page - 1) * in.PageSize
	limit := in.PageSize

	// 获取分站对应的所有供应商代码
	providerCodes, err := l.svcCtx.DAOManager.OperatorGameProvider.GetProviderCodesByOpCode(l.ctx, in.OpCode)
	if err != nil {
		l.Errorf("GetProviderCodesByOpCode failed: %v", err)
		return nil, err
	}

	// 获取供应商信息
	providers, total, err := l.svcCtx.DAOManager.GameProvider.GetPublishedGameProviderList(l.ctx, providerCodes, int64(offset), int64(limit))
	if err != nil {
		l.Errorf("GetPublishedGameProviderList failed: %v", err)
		return nil, err
	}

	// 转换为proto消息
	items := make([]*platform_game.PublishedGameProviderInfo, 0, len(providers))
	for _, provider := range providers {
		items = append(items, &platform_game.PublishedGameProviderInfo{
			Id:           provider.ID,
			ProviderCode: provider.SourceProviderCode,
			ChannelCode:  provider.ChannelCode,
			LogoUrl:      provider.LogoURL,
			SortNo:       provider.SortNo,
			Status:       int32(provider.Status),
			CreatedAt:    provider.CreatedAt.Unix() * 1000, // 转换为毫秒时间戳
			UpdatedAt:    provider.UpdatedAt.Unix() * 1000,
		})
	}

	return &platform_game.GetPublishedGameProviderListResp{
		Items:    items,
		Total:    int64(total),
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil
}
