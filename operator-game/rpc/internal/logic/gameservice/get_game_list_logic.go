package gameservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-game/rpc/internal/dao"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/internal/utils"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameListLogic {
	return &GetGameListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取游戏列表
func (l *GetGameListLogic) GetGameList(in *operator_game.GetGameListRequest) (*operator_game.GetGameListResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameList] DAO Manager not available")
		return &operator_game.GetGameListResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	page, pageSize := utils.HandlePage(int64(in.Page), int64(in.PageSize))
	offset := (page - 1) * pageSize

	opts := []dao.GameListOption{
		dao.WithGameOffset(offset),
		dao.WithGameLimit(pageSize),
	}

	// 从上下文获取 OperatorCode
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode != "" {
		opts = append(opts, dao.WithGameOpCode(operatorCode))
	}

	if in.GameCode != "" {
		opts = append(opts, dao.WithGameCode(in.GameCode))
	}
	if in.Name != "" {
		opts = append(opts, dao.WithGameName(in.Name))
	}
	if in.CategoryCode != "" {
		opts = append(opts, dao.WithGameCategoryCode(in.CategoryCode))
	}
	if in.ProviderCode != "" {
		opts = append(opts, dao.WithGameProviderCode(in.ProviderCode))
	}
	if in.ChannelCode != "" {
		opts = append(opts, dao.WithGameChannelCode(in.ChannelCode))
	}
	if in.Status > 0 {
		opts = append(opts, dao.WithGameStatus(int64(in.Status)))
	}
	games, err := l.svcCtx.DAOManager.Game.GetGameList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameList] query failed: %v", err)
		return &operator_game.GetGameListResp{
			Code:    500,
			Message: "failed to get game list: " + err.Error(),
		}, nil
	}

	total, err := l.svcCtx.DAOManager.Game.CountGameList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameList] count failed: %v", err)
		return &operator_game.GetGameListResp{
			Code:    500,
			Message: "failed to count game list: " + err.Error(),
		}, nil
	}

	var items []*operator_game.GameInfo
	for _, g := range games {
		// 查询对应currency
		currencyCodeList := make([]string, 0)
		for _, c := range g.CurrencyList {
			currencyCodeList = append(currencyCodeList, c.Code)
		}
		items = append(items, &operator_game.GameInfo{
			Id:               g.ID,
			SourceId:         g.SourceID,
			GameCode:         g.GameCode,
			Name:             g.Name,
			Status:           int32(g.Status),
			CategoryCode:     g.CategoryCode,
			ProviderCode:     g.ProviderCode,
			ChannelCode:      g.ChannelCode,
			CurrencyCodeList: currencyCodeList,
			ProviderKey:      g.ProviderKey,
			ImageUrl:         g.ImageURL,
			SortNo:           g.SortNo,
			SupportsEmbed:    g.SupportsEmbed,
			SupportsRedirect: g.SupportsRedirect,
			CreatedAt:        g.CreatedAt.UnixMilli(),
			UpdatedAt:        g.UpdatedAt.UnixMilli(),
		})
	}

	return &operator_game.GetGameListResp{
		Code:     0,
		Message:  "ok",
		Items:    items,
		Total:    int64(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
