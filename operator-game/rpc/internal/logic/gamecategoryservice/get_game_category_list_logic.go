package gamecategoryservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/dao"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/internal/utils"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameCategoryListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameCategoryListLogic {
	return &GetGameCategoryListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取游戏分类列表
func (l *GetGameCategoryListLogic) GetGameCategoryList(in *operator_game.GetGameCategoryListRequest) (*operator_game.GetGameCategoryListResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameCategoryList] DAO Manager not available")
		return &operator_game.GetGameCategoryListResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	page, pageSize := utils.HandlePage(int64(in.Page), int64(in.PageSize))
	offset := (page - 1) * pageSize

	opts := []dao.GameCategoryListOption{
		dao.WithGameCategoryOffset(offset),
		dao.WithGameCategoryLimit(pageSize),
	}

	if in.CategoryCode != "" {
		opts = append(opts, dao.WithCategoryCode(in.CategoryCode))
	}
	if in.Status > 0 {
		opts = append(opts, dao.WithGameCategoryStatus(int64(in.Status)))
	}

	gcs, err := l.svcCtx.DAOManager.GameCategory.GetGameCategoryList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameCategoryList] query failed: %v", err)
		return &operator_game.GetGameCategoryListResp{
			Code:    500,
			Message: "failed to get game category list: " + err.Error(),
		}, nil
	}

	total, err := l.svcCtx.DAOManager.GameCategory.CountGameCategoryList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameCategoryList] count failed: %v", err)
		return &operator_game.GetGameCategoryListResp{
			Code:    500,
			Message: "failed to count game category list: " + err.Error(),
		}, nil
	}

	var items []*operator_game.GameCategoryInfo
	for _, gc := range gcs {
		items = append(items, &operator_game.GameCategoryInfo{
			Id:           gc.ID,
			CategoryCode: gc.CategoryCode,
			SortNo:       int32(gc.SortNo),
			Status:       int32(gc.Status),
			CreatedAt:    gc.CreatedAt.UnixMilli(),
			UpdatedAt:    gc.UpdatedAt.UnixMilli(),
		})
	}

	return &operator_game.GetGameCategoryListResp{
		Code:    0,
		Message: "ok",
		Items:   items,
		Total:   int64(total),
	}, nil
}
