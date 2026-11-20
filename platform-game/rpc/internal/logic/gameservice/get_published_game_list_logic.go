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
	for _, game := range games {
		// 转换category_id为category_code
		categoryRecord, _ := l.svcCtx.DAOManager.GameCategory.GetGameCategoryBySourceId(l.ctx, game.CategoryID)
		categoryCode := ""
		if categoryRecord != nil {
			categoryCode = categoryRecord.SourceCategoryCode
		}
		providerRecord, _ := l.svcCtx.DAOManager.GameProvider.GetGameProviderBySourceId(l.ctx, game.ProviderID)
		providerCode := ""
		if providerRecord != nil {
			providerCode = providerRecord.SourceProviderCode
		}

		channelRecord, _ := l.svcCtx.DAOManager.GameChannel.GetGameChannelBySourceId(l.ctx, game.ChannelID)
		channelCode := ""
		if channelRecord != nil {
			channelCode = channelRecord.SourceChannelCode
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
		GameItems: items,
		Total:     int64(total),
		Page:      in.Page,
		PageSize:  in.PageSize,
	}, nil
}
