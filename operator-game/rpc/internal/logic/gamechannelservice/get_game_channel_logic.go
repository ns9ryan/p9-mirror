package gamechannelservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGameChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGameChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGameChannelLogic {
	return &GetGameChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个游戏渠道
func (l *GetGameChannelLogic) GetGameChannel(in *operator_game.GetGameChannelRequest) (*operator_game.GetGameChannelResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC GetGameChannel] DAO Manager not available")
		return &operator_game.GetGameChannelResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	gc, err := l.svcCtx.DAOManager.GameChannel.GetGameChannelByID(l.ctx, in.GetId())
	if err != nil {
		l.Errorf("[RPC GetGameChannel] query failed: %v", err)
		return &operator_game.GetGameChannelResp{
			Code:    500,
			Message: "failed to get game channel: " + err.Error(),
		}, nil
	}

	return &operator_game.GetGameChannelResp{
		Code:    0,
		Message: "ok",
		Data: &operator_game.GameChannelInfo{
			Id:          gc.ID,
			ChannelCode: gc.ChannelCode,
			SortNo:      int32(gc.SortNo),
			LoadType:    int32(gc.LoadType),
			Status:      int32(gc.Status),
			CreatedAt:   gc.CreatedAt.UnixMilli(),
			UpdatedAt:   gc.UpdatedAt.UnixMilli(),
		},
	}, nil
}
