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

type GameUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameUpdateLogic {
	return &GameUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameUpdateLogic) GameUpdate(req *types.GameUpdateReq) (resp *types.GameResp, err error) {
	pbReq := &pb.UpdateGameRequest{
		Id:     req.ID,
		Name:   req.Name,
		SortNo: int32(req.SortNo),
		Status: int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameServiceClient().UpdateGame(l.ctx, pbReq)
	if err != nil {
		l.Errorf("UpdateGame failed: %v", err)
		return nil, err
	}

	return logic.GameProtoToResponse(l.ctx, pbResp.Data), nil
}
