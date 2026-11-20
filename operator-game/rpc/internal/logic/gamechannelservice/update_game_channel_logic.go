package gamechannelservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-game/rpc/internal/svc"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGameChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGameChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGameChannelLogic {
	return &UpdateGameChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新游戏渠道
func (l *UpdateGameChannelLogic) UpdateGameChannel(in *operator_game.UpdateGameChannelRequest) (*operator_game.UpdateGameChannelResp, error) {
	if l.svcCtx == nil || l.svcCtx.DAOManager == nil {
		l.Errorf("[RPC UpdateGameChannel] DAO Manager not available")
		return &operator_game.UpdateGameChannelResp{
			Code:    500,
			Message: "DAO Manager not available",
		}, nil
	}

	gc, err := l.svcCtx.DAOManager.GameChannel.UpdateGameChannel(l.ctx, in.Id, int64(in.SortNo), int64(in.Status))
	if err != nil {
		l.Errorf("[RPC UpdateGameChannel] update failed: %v", err)
		return &operator_game.UpdateGameChannelResp{
			Code:    500,
			Message: "failed to update game channel: " + err.Error(),
		}, nil
	}

	return &operator_game.UpdateGameChannelResp{
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
