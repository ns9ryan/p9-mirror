package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameRebateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGameRebateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameRebateLogic {
	return &GameRebateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 游戏返水
func (l *GameRebateLogic) GameRebate(in *gamepb.GameRebateRequest) (*gamepb.GameRebateResponse, error) {
	// todo: add your logic here and delete this line

	return &gamepb.GameRebateResponse{}, nil
}
