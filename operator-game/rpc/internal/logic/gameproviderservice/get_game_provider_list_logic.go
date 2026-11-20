package gameproviderservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-game/rpc/internal/dao"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/internal/utils"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameProviderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameProviderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameProviderListLogic {
	return &GetGameProviderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取游戏供应商列表
func (l *GetGameProviderListLogic) GetGameProviderList(in *operator_game.GetGameProviderListRequest) (*operator_game.GetGameProviderListResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameProviderList] DAO Manager not available")
		return &operator_game.GetGameProviderListResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	page, pageSize := utils.HandlePage(int64(in.Page), int64(in.PageSize))
	offset := (page - 1) * pageSize

	opts := []dao.GameProviderListOption{
		dao.WithGameProviderOffset(offset),
		dao.WithGameProviderLimit(pageSize),
	}

	// 从上下文获取 OperatorCode
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode != "" {
		opts = append(opts, dao.WithProviderOpCode(operatorCode))
	}

	if in.ProviderCode != "" {
		opts = append(opts, dao.WithProviderCode(in.ProviderCode))
	}
	if in.Status > 0 {
		opts = append(opts, dao.WithGameProviderStatus(int64(in.Status)))
	}

	gps, err := l.svcCtx.DAOManager.GameProvider.GetGameProviderList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameProviderList] query failed: %v", err)
		return &operator_game.GetGameProviderListResp{
			Code:    500,
			Message: "failed to get game provider list: " + err.Error(),
		}, nil
	}

	total, err := l.svcCtx.DAOManager.GameProvider.CountGameProviderList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameProviderList] count failed: %v", err)
		return &operator_game.GetGameProviderListResp{
			Code:    500,
			Message: "failed to count game provider list: " + err.Error(),
		}, nil
	}

	var items []*operator_game.ProviderInfo
	for _, gp := range gps {
		items = append(items, &operator_game.ProviderInfo{
			Id:           gp.ID,
			ProviderCode: gp.ProviderCode,
			ChannelCode:  gp.ChannelCode,
			LogoUrl:      gp.LogoURL,
			SortNo:       int32(gp.SortNo),
			Status:       int32(gp.Status),
			CreatedAt:    gp.CreatedAt.UnixMilli(),
			UpdatedAt:    gp.UpdatedAt.UnixMilli(),
		})
	}

	return &operator_game.GetGameProviderListResp{
		Code:    0,
		Message: "ok",
		Items:   items,
		Total:   int64(total),
	}, nil
}
