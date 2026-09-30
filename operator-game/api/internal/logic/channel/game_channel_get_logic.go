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

type GameChannelGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameChannelGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameChannelGetLogic {
	return &GameChannelGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameChannelGetLogic) GameChannelGet(req *types.GameChannelGetReq) (resp *types.GameChannelResp, err error) {
	pbReq := &pb.GetGameChannelRequest{
		Id: req.ID,
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameChannelServiceClient().GetGameChannel(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGameChannel failed: %v", err)
		return nil, err
	}

	return logic.ChannelProtoToResponse(l.ctx, pbResp.Data), nil
}
