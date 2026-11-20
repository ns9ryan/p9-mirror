package logic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-game/api/internal/constant"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	"oa.98ent.com/p9/operator-game/api/internal/utils"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"
)

// CategoryProtoToResponse 将 proto GameCategoryInfo 转换为 API GameCategoryResp
func CategoryProtoToResponse(ctx context.Context, category *pb.GameCategoryInfo) *types.GameCategoryResp {
	nameKey := fmt.Sprintf("%s.%s.name", constant.CategoryBiz, category.CategoryCode)
	// 调用 TG 进行翻译
	name := utils.TGPlatformGame(ctx, nameKey)
	return &types.GameCategoryResp{
		ID:           category.Id,
		CategoryCode: category.CategoryCode,
		Name:         name, // 保留 name 字段，使用 name_key 的值
		SortNo:       category.SortNo,
		Status:       category.Status,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
	}
}

// ChannelProtoToResponse 将 proto GameChannelInfo 转换为 API GameChannelResp
func ChannelProtoToResponse(ctx context.Context, channel *pb.GameChannelInfo) *types.GameChannelResp {
	nameKey := fmt.Sprintf("%s.%s.name", constant.ChannelBiz, channel.ChannelCode)
	name := utils.TGPlatformGame(ctx, nameKey)
	return &types.GameChannelResp{
		ID:          channel.Id,
		ChannelCode: channel.ChannelCode,
		Name:        name, // 保留 name 字段
		SortNo:      channel.SortNo,
		LoadType:    channel.LoadType,
		Status:      channel.Status,
		CreatedAt:   channel.CreatedAt,
		UpdatedAt:   channel.UpdatedAt,
	}
}

// ProviderProtoToResponse 将 proto ProviderInfo 转换为 API GameProviderResp
func ProviderProtoToResponse(ctx context.Context, provider *pb.ProviderInfo) *types.GameProviderResp {
	providerNameKey := fmt.Sprintf("%s.%s.name", constant.ProviderBiz, provider.ProviderCode)
	providerName := utils.TGPlatformGame(ctx, providerNameKey)
	channelName := ""
	if provider.ChannelCode != "" {
		channelNameKey := fmt.Sprintf("%s.%s.name", constant.ChannelBiz, provider.ChannelCode)
		channelName = utils.TGPlatformGame(ctx, channelNameKey)
	}
	return &types.GameProviderResp{
		ID:           provider.Id,
		ProviderCode: provider.ProviderCode,
		ChannelCode:  provider.ChannelCode,
		ProviderName: providerName,
		ChannelName:  channelName,
		LogoUrl:      provider.LogoUrl,
		SortNo:       provider.SortNo,
		Status:       provider.Status,
		CreatedAt:    provider.CreatedAt,
		UpdatedAt:    provider.UpdatedAt,
	}
}

// GameProtoToResponse 将 proto GameInfo 转换为 API GameResp
func GameProtoToResponse(ctx context.Context, game *pb.GameInfo) *types.GameResp {
	categoryNameKey := fmt.Sprintf("%s.%s.name", constant.CategoryBiz, game.CategoryCode)
	providerNameKey := fmt.Sprintf("%s.%s.name", constant.ProviderBiz, game.ProviderCode)
	channelNameKey := fmt.Sprintf("%s.%s.name", constant.ChannelBiz, game.ChannelCode)
	categoryName := utils.TGPlatformGame(ctx, categoryNameKey)
	providerName := utils.TGPlatformGame(ctx, providerNameKey)
	channelName := utils.TGPlatformGame(ctx, channelNameKey)
	return &types.GameResp{
		ID:               game.Id,
		SourceID:         game.SourceId,
		GameCode:         game.GameCode,
		Name:             game.Name,
		Status:           game.Status,
		CategoryCode:     game.CategoryCode,
		ProviderCode:     game.ProviderCode,
		ChannelCode:      game.ChannelCode,
		CategoryName:     categoryName,
		ProviderName:     providerName,
		ChannelName:      channelName,
		ProviderKey:      game.ProviderKey,
		CurrencyCodeList: game.CurrencyCodeList,
		ImageUrl:         game.ImageUrl,
		SortNo:           game.SortNo,
		SupportsEmbed:    game.SupportsEmbed,
		SupportsRedirect: game.SupportsRedirect,
		CreatedAt:        game.CreatedAt,
		UpdatedAt:        game.UpdatedAt,
	}
}
