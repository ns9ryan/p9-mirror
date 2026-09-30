// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package channel

import (
	"context"

	"oa.98ent.com/p9/operator-game/api/internal/logic"
	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameChannelUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameChannelUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameChannelUpdateLogic {
	return &GameChannelUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameChannelUpdateLogic) GameChannelUpdate(req *types.GameChannelUpdateReq) (resp *types.GameChannelResp, err error) {
	pbReq := &pb.UpdateGameChannelRequest{
		Id:     req.ID,
		SortNo: req.SortNo,
		Status: int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameChannelServiceClient().UpdateGameChannel(l.ctx, pbReq)
	if err != nil {
		l.Errorf("UpdateGameChannel failed: %v", err)
		return nil, err
	}

	return logic.ChannelProtoToResponse(l.ctx, pbResp.Data), nil
}
