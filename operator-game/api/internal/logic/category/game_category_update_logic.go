// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package category

import (
	"context"

	"oa.98ent.com/p9/operator-game/api/internal/logic"
	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameCategoryUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameCategoryUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameCategoryUpdateLogic {
	return &GameCategoryUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameCategoryUpdateLogic) GameCategoryUpdate(req *types.GameCategoryUpdateReq) (resp *types.GameCategoryResp, err error) {
	pbReq := &pb.UpdateGameCategoryRequest{
		Id:     req.ID,
		SortNo: req.SortNo,
		Status: int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameCategoryServiceClient().UpdateGameCategory(l.ctx, pbReq)
	if err != nil {
		l.Errorf("UpdateGameCategory failed: %v", err)
		return nil, err
	}

	return logic.CategoryProtoToResponse(l.ctx, pbResp.Data), nil
}
