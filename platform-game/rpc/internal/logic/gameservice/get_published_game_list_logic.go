package gameservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublishedGameListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPublishedGameListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublishedGameListLogic {
	return &GetPublishedGameListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取已发布的游戏列表（供分站同步）
func (l *GetPublishedGameListLogic) GetPublishedGameList(in *platform_game.GetPublishedGameListRequest) (*platform_game.GetPublishedGameListResp, error) {
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

	// 获取分站对应的所有游戏ID
	gameCodes, err := l.svcCtx.DAOManager.OperatorGame.GetGameCodesByOpCode(l.ctx, in.OpCode)
	if err != nil {
		l.Errorf("GetGameCodesByOpCode failed: %v", err)
		return nil, err
	}

	// 获取游戏信息
	games, total, err := l.svcCtx.DAOManager.Game.GetPublishedGameList(l.ctx, gameCodes, int64(offset), int64(limit))
	if err != nil {
		l.Errorf("GetPublishedGameList failed: %v", err)
		return nil, err
	}

	// 转换为proto消息
	items := make([]*platform_game.PublishedGameInfo, 0, len(games))
	categoryItems := make([]*platform_game.PublishedGameCategoryInfo, 0)
	providerItems := make([]*platform_game.PublishedGameProviderInfo, 0)
	channelItems := make([]*platform_game.PublishedGameChannelInfo, 0)
	for _, game := range games {
		// 转换category_id为category_code
		categoryRecord, _ := l.svcCtx.DAOManager.GameCategory.GetGameCategoryBySourceId(l.ctx, game.CategoryID)
		categoryCode := ""
		if categoryRecord != nil {
			categoryCode = categoryRecord.SourceCategoryCode
			categoryItems = append(categoryItems, &platform_game.PublishedGameCategoryInfo{
				Id:           categoryRecord.SourceID,
				CategoryCode: categoryCode,
				SortNo:       categoryRecord.SortNo,
				Status:       int32(categoryRecord.Status),
			})
		}
		providerRecord, _ := l.svcCtx.DAOManager.GameProvider.GetGameProviderBySourceId(l.ctx, game.ProviderID)
		providerCode := ""
		if providerRecord != nil {
			providerCode = providerRecord.SourceProviderCode
			providerItems = append(providerItems, &platform_game.PublishedGameProviderInfo{
				Id:           providerRecord.SourceID,
				ProviderCode: providerCode,
				ChannelCode:  providerRecord.ChannelCode,
				LogoUrl:      providerRecord.LogoURL,
				SortNo:       providerRecord.SortNo,
				Status:       int32(providerRecord.Status),
			})
		}

		channelRecord, _ := l.svcCtx.DAOManager.GameChannel.GetGameChannelBySourceId(l.ctx, game.ChannelID)
		channelCode := ""
		if channelRecord != nil {
			channelCode = channelRecord.SourceChannelCode
			channelItems = append(channelItems, &platform_game.PublishedGameChannelInfo{
				Id:          channelRecord.SourceID,
				ChannelCode: channelCode,
				LoadType:    channelRecord.LoadType,
				SortNo:      channelRecord.SortNo,
				Status:      int32(channelRecord.Status),
			})
		}

		currencyCodeList, _ := l.svcCtx.DAOManager.GameCurrency.GetCurrencyCodesByGameCode(l.ctx, game.SourceGameCode)

		items = append(items, &platform_game.PublishedGameInfo{
			Id:               game.ID,
			SourceId:         game.SourceID,
			GameCode:         game.SourceGameCode,
			CategoryCode:     categoryCode,
			ProviderCode:     providerCode,
			ChannelCode:      channelCode,
			CurrencyCodeList: currencyCodeList,
			Name:             game.Name,
			ProviderKey:      game.ProviderKey,
			ImageUrl:         game.ImageURL,
			SortNo:           game.SortNo,
			SupportsEmbed:    game.SupportsEmbed,
			SupportsRedirect: game.SupportsRedirect,
			Status:           int32(game.Status),
			CreatedAt:        game.CreatedAt.Unix() * 1000, // 转换为毫秒时间戳
			UpdatedAt:        game.UpdatedAt.Unix() * 1000,
		})
	}

	return &platform_game.GetPublishedGameListResp{
		GameItems:     items,
		CategoryItems: categoryItems,
		ProviderItems: providerItems,
		ChannelItems:  channelItems,
		Total:         int64(total),
		Page:          in.Page,
		PageSize:      in.PageSize,
	}, nil
}
