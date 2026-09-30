package gamecategoryservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGameCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGameCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGameCategoryLogic {
	return &UpdateGameCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新游戏分类
func (l *UpdateGameCategoryLogic) UpdateGameCategory(in *operator_game.UpdateGameCategoryRequest) (*operator_game.UpdateGameCategoryResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC UpdateGameCategory] DAO Manager not available")
		return &operator_game.UpdateGameCategoryResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	gc, err := l.svcCtx.DAOManager.GameCategory.UpdateGameCategory(l.ctx, in.Id, int64(in.SortNo), int64(in.Status))
	if err != nil {
		l.Errorf("[RPC UpdateGameCategory] update failed: %v", err)
		return &operator_game.UpdateGameCategoryResp{
			Code:    500,
			Message: "failed to update game category: " + err.Error(),
		}, nil
	}

	return &operator_game.UpdateGameCategoryResp{
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
