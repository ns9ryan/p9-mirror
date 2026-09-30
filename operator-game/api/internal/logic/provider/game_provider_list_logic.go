// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package provider

import (
	"context"

	"oa.98ent.com/p9/operator-game/api/internal/logic"
	"oa.98ent.com/p9/operator-game/api/internal/svc"
	"oa.98ent.com/p9/operator-game/api/internal/types"
	pb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameProviderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameProviderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameProviderListLogic {
	return &GameProviderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameProviderListLogic) GameProviderList(req *types.GameProviderListReq) (resp *types.GameProviderListResp, err error) {
	pbReq := &pb.GetGameProviderListRequest{
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
		ProviderCode: req.ProviderCode,
		Status:       int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameProviderServiceClient().GetGameProviderList(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGameProviderList failed: %v", err)
		return nil, err
	}

	var respList []types.GameProviderResp
	for _, item := range pbResp.Items {
		respList = append(respList, *logic.ProviderProtoToResponse(l.ctx, item))
	}

	return &types.GameProviderListResp{
		List:  respList,
		Total: pbResp.Total,
	}, nil
}
