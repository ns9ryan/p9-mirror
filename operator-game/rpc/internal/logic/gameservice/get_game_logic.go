package gameservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameLogic {
	return &GetGameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个游戏详情
func (l *GetGameLogic) GetGame(in *operator_game.GetGameRequest) (*operator_game.GetGameResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGame] DAO Manager not available")
		return &operator_game.GetGameResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	g, err := l.svcCtx.DAOManager.Game.GetGameByID(l.ctx, in.GetId())
	if err != nil {
		l.Errorf("[RPC GetGame] query failed: %v", err)
		return &operator_game.GetGameResp{
			Code:    500,
			Message: "failed to get game: " + err.Error(),
		}, nil
	}

	return &operator_game.GetGameResp{
		Code:    0,
		Message: "ok",
		Data: &operator_game.GameInfo{
			Id:               g.ID,
			SourceId:         g.SourceID,
			GameCode:         g.GameCode,
			Name:             g.Name,
			Status:           int32(g.Status),
			CategoryCode:     g.CategoryCode,
			ProviderCode:     g.ProviderCode,
			ChannelCode:      g.ChannelCode,
			ProviderKey:      g.ProviderKey,
			ImageUrl:         g.ImageURL,
			SortNo:           g.SortNo,
			SupportsEmbed:    g.SupportsEmbed,
			SupportsRedirect: g.SupportsRedirect,
			CreatedAt:        g.CreatedAt.UnixMilli(),
			UpdatedAt:        g.UpdatedAt.UnixMilli(),
		},
	}, nil
}
