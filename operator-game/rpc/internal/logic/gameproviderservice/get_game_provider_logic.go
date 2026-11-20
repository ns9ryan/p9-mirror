package gameproviderservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameProviderLogic {
	return &GetGameProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个游戏供应商
func (l *GetGameProviderLogic) GetGameProvider(in *operator_game.GetGameProviderRequest) (*operator_game.GetGameProviderResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameProvider] DAO Manager not available")
		return &operator_game.GetGameProviderResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	gp, err := l.svcCtx.DAOManager.GameProvider.GetGameProviderByID(l.ctx, in.GetId())
	if err != nil {
		l.Errorf("[RPC GetGameProvider] query failed: %v", err)
		return &operator_game.GetGameProviderResp{
			Code:    500,
			Message: "failed to get game provider: " + err.Error(),
		}, nil
	}

	return &operator_game.GetGameProviderResp{
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
