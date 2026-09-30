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

type GameListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameListLogic {
	return &GameListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameListLogic) GameList(req *types.GameListReq) (resp *types.GameListResp, err error) {
	pbReq := &pb.GetGameListRequest{
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
		GameCode:     req.GameCode,
		Name:         req.Name,
		CategoryCode: req.CategoryCode,
		ProviderCode: req.ProviderCode,
		ChannelCode:  req.ChannelCode,
		Status:       int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameServiceClient().GetGameList(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGameList failed: %v", err)
		return nil, err
	}

	var respList []types.GameResp
	for _, item := range pbResp.Items {
		respList = append(respList, *logic.GameProtoToResponse(l.ctx, item))
	}

	return &types.GameListResp{
		List:  respList,
		Total: pbResp.Total,
	}, nil
}
