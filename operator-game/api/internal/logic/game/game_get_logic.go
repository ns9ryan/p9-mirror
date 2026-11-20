// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package game

import (
	"context"

	"oa.98ent.com/p9/operator-game/api/internal/logic"
	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameGetLogic {
	return &GameGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameGetLogic) GameGet(req *types.GameGetReq) (resp *types.GameResp, err error) {
	pbReq := &pb.GetGameRequest{
		Id: req.ID,
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameServiceClient().GetGame(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGame failed: %v", err)
		return nil, err
	}

	return logic.GameProtoToResponse(l.ctx, pbResp.Data), nil
}
