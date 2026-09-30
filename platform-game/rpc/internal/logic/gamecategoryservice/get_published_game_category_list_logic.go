package gamecategoryservicelogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/platform-game/rpc/internal/svc"
	"oa.98ent.com/p9/platform-game/rpc/pb/platform_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublishedGameCategoryListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPublishedGameCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublishedGameCategoryListLogic {
	return &GetPublishedGameCategoryListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取已发布的游戏分类列表（供分站同步）
func (l *GetPublishedGameCategoryListLogic) GetPublishedGameCategoryList(in *platform_game.GetPublishedGameCategoryListRequest) (*platform_game.GetPublishedGameCategoryListResp, error) {
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

	// 获取分站对应的所有分类代码
	categoryCodes, err := l.svcCtx.DAOManager.OperatorGameCategory.GetCategoryCodesByOpCode(l.ctx, in.OpCode)
	if err != nil {
		l.Errorf("GetCategoryCodesByOpCode failed: %v", err)
		return nil, err
	}

	// 获取分类信息
	categories, total, err := l.svcCtx.DAOManager.GameCategory.GetPublishedGameCategoryList(l.ctx, categoryCodes, int64(offset), int64(limit))
	if err != nil {
		l.Errorf("GetPublishedGameCategoryList failed: %v", err)
		return nil, err
	}

	// 转换为proto消息
	items := make([]*platform_game.PublishedGameCategoryInfo, 0, len(categories))
	for _, cat := range categories {
		items = append(items, &platform_game.PublishedGameCategoryInfo{
			Id:           cat.ID,
			CategoryCode: cat.SourceCategoryCode,
			SortNo:       cat.SortNo,
			Status:       int32(cat.Status),
			CreatedAt:    cat.CreatedAt.Unix() * 1000, // 转换为毫秒时间戳
			UpdatedAt:    cat.UpdatedAt.Unix() * 1000,
		})
	}

	return &platform_game.GetPublishedGameCategoryListResp{
		Items:    items,
		Total:    int64(total),
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil
}
