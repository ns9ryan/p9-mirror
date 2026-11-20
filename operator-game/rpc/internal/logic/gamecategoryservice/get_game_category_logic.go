package gamecategoryservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameCategoryLogic {
	return &GetGameCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个游戏分类
func (l *GetGameCategoryLogic) GetGameCategory(in *operator_game.GetGameCategoryRequest) (*operator_game.GetGameCategoryResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameCategory] DAO Manager not available")
		return &operator_game.GetGameCategoryResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	gc, err := l.svcCtx.DAOManager.GameCategory.GetGameCategoryByID(l.ctx, in.GetId())
	if err != nil {
		l.Errorf("[RPC GetGameCategory] query failed: %v", err)
		return &operator_game.GetGameCategoryResp{
			Code:    500,
			Message: "failed to get game category: " + err.Error(),
		}, nil
	}

	return &operator_game.GetGameCategoryResp{
		Code:    0,
		Message: "ok",
		Data: &operator_game.GameCategoryInfo{
			Id:           gc.ID,
			CategoryCode: gc.CategoryCode,
			SortNo:       int32(gc.SortNo),
			Status:       int32(gc.Status),
			CreatedAt:    gc.CreatedAt.UnixMilli(),
			UpdatedAt:    gc.UpdatedAt.UnixMilli(),
		},
	}, nil
}
