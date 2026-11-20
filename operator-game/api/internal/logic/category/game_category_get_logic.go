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

type GameCategoryGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameCategoryGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameCategoryGetLogic {
	return &GameCategoryGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameCategoryGetLogic) GameCategoryGet(req *types.GameCategoryGetReq) (resp *types.GameCategoryResp, err error) {
	pb := &pb.GetGameCategoryRequest{
		Id: req.ID,
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameCategoryServiceClient().GetGameCategory(l.ctx, pb)
	if err != nil {
		l.Errorf("GetGameCategory failed: %v", err)
		return nil, err
	}

	return logic.CategoryProtoToResponse(l.ctx, pbResp.Data), nil
}
