package gameproviderservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGameProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGameProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGameProviderLogic {
	return &UpdateGameProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新游戏供应商
func (l *UpdateGameProviderLogic) UpdateGameProvider(in *operator_game.UpdateGameProviderRequest) (*operator_game.UpdateGameProviderResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC UpdateGameProvider] DAO Manager not available")
		return &operator_game.UpdateGameProviderResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	gp, err := l.svcCtx.DAOManager.GameProvider.UpdateGameProvider(l.ctx, in.Id, int64(in.SortNo), int64(in.Status))
	if err != nil {
		l.Errorf("[RPC UpdateGameProvider] update failed: %v", err)
		return &operator_game.UpdateGameProviderResp{
			Code:    500,
			Message: "failed to update game provider: " + err.Error(),
		}, nil
	}

	return &operator_game.UpdateGameProviderResp{
		Code:    0,
		Message: "ok",
		Data: &operator_game.ProviderInfo{
			Id:           gp.ID,
			ProviderCode: gp.ProviderCode,
			LogoUrl:      gp.LogoURL,
			SortNo:       int32(gp.SortNo),
			Status:       int32(gp.Status),
			CreatedAt:    gp.CreatedAt.UnixMilli(),
			UpdatedAt:    gp.UpdatedAt.UnixMilli(),
		},
	}, nil
}
