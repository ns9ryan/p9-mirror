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

type GameProviderUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameProviderUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameProviderUpdateLogic {
	return &GameProviderUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameProviderUpdateLogic) GameProviderUpdate(req *types.GameProviderUpdateReq) (resp *types.GameProviderResp, err error) {
	pbReq := &pb.UpdateGameProviderRequest{
		Id:     req.ID,
		SortNo: req.SortNo,
		Status: int32(req.Status),
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameProviderServiceClient().UpdateGameProvider(l.ctx, pbReq)
	if err != nil {
		l.Errorf("UpdateGameProvider failed: %v", err)
		return nil, err
	}

	return logic.ProviderProtoToResponse(l.ctx, pbResp.Data), nil
}
