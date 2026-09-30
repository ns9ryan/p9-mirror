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

type GameProviderGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGameProviderGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameProviderGetLogic {
	return &GameProviderGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GameProviderGetLogic) GameProviderGet(req *types.GameProviderGetReq) (resp *types.GameProviderResp, err error) {
	pbReq := &pb.GetGameProviderRequest{
		Id: req.ID,
	}

	pbResp, err := l.svcCtx.OperatorGameGrpcClient.GetGameProviderServiceClient().GetGameProvider(l.ctx, pbReq)
	if err != nil {
		l.Errorf("GetGameProvider failed: %v", err)
		return nil, err
	}

	return logic.ProviderProtoToResponse(l.ctx, pbResp.Data), nil
}
