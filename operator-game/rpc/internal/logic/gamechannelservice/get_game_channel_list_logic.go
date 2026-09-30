package gamechannelservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-game/rpc/internal/dao"
	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/internal/utils"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameChannelListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameChannelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameChannelListLogic {
	return &GetGameChannelListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取游戏渠道列表
func (l *GetGameChannelListLogic) GetGameChannelList(in *operator_game.GetGameChannelListRequest) (*operator_game.GetGameChannelListResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameChannelList] DAO Manager not available")
		return &operator_game.GetGameChannelListResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	page, pageSize := utils.HandlePage(int64(in.Page), int64(in.PageSize))
	offset := (page - 1) * pageSize

	opts := []dao.GameChannelListOption{
		dao.WithGameChannelOffset(offset),
		dao.WithGameChannelLimit(pageSize),
	}

	// 从上下文获取 OperatorCode
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode != "" {
		opts = append(opts, dao.WithChannelOpCode(operatorCode))
	}

	if in.ChannelCode != "" {
		opts = append(opts, dao.WithChannelCode(in.ChannelCode))
	}
	if in.Status > 0 {
		opts = append(opts, dao.WithGameChannelStatus(int64(in.Status)))
	}

	gcs, err := l.svcCtx.DAOManager.GameChannel.GetGameChannelList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameChannelList] query failed: %v", err)
		return &operator_game.GetGameChannelListResp{
			Code:    500,
			Message: "failed to get game channel list: " + err.Error(),
		}, nil
	}

	total, err := l.svcCtx.DAOManager.GameChannel.CountGameChannelList(l.ctx, opts...)
	if err != nil {
		l.Errorf("[RPC GetGameChannelList] count failed: %v", err)
		return &operator_game.GetGameChannelListResp{
			Code:    500,
			Message: "failed to count game channel list: " + err.Error(),
		}, nil
	}

	var items []*operator_game.GameChannelInfo
	for _, gc := range gcs {
		// 获取渠道对应的游戏厂商数量
		providerCount, err := l.svcCtx.DAOManager.GameProvider.CountGameProviderByChannel(l.ctx, gc.ChannelCode)
		if err != nil {
			l.Errorf("[RPC GetGameChannelList] count providers failed: %v", err)
			providerCount = 0
		}
		// 获取渠道对应的游戏数量
		gameCount, err := l.svcCtx.DAOManager.Game.CountGameByChannel(l.ctx, gc.ChannelCode)
		if err != nil {
			l.Errorf("[RPC GetGameChannelList] count games failed: %v", err)
			gameCount = 0
		}
		items = append(items, &operator_game.GameChannelInfo{
			Id:            gc.ID,
			ChannelCode:   gc.ChannelCode,
			ProviderCount: int32(providerCount),
			GameCount:     int32(gameCount),
			SortNo:        int32(gc.SortNo),
			LoadType:      int32(gc.LoadType),
			Status:        int32(gc.Status),
			CreatedAt:     gc.CreatedAt.UnixMilli(),
			UpdatedAt:     gc.UpdatedAt.UnixMilli(),
		})
	}

	return &operator_game.GetGameChannelListResp{
		Code:    0,
		Message: "ok",
		Items:   items,
		Total:   int64(total),
	}, nil
}
