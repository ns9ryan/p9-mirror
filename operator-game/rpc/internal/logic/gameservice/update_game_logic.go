package gameservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGameLogic {
	return &UpdateGameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新游戏
func (l *UpdateGameLogic) UpdateGame(in *operator_game.UpdateGameRequest) (*operator_game.UpdateGameResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC UpdateGame] DAO Manager not available")
		return &operator_game.UpdateGameResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	g, err := l.svcCtx.DAOManager.Game.UpdateGame(l.ctx, in.Id, in.Name, int64(in.SortNo), int64(in.Status))
	if err != nil {
		l.Errorf("[RPC UpdateGame] update failed: %v", err)
		return &operator_game.UpdateGameResp{
			Code:    500,
			Message: "failed to update game: " + err.Error(),
		}, nil
	}

	return &operator_game.UpdateGameResp{
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
