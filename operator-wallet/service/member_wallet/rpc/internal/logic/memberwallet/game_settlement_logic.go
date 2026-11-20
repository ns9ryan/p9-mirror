package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGameSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameSettlementLogic {
	return &GameSettlementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 游戏结算
func (l *GameSettlementLogic) GameSettlement(in *gamepb.GameSettlementRequest) (*gamepb.GameSettlementResponse, error) {
	// todo: add your logic here and delete this line

	return &gamepb.GameSettlementResponse{}, nil
}
