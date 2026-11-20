package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameSettlementRollbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGameSettlementRollbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameSettlementRollbackLogic {
	return &GameSettlementRollbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 游戏结算回滚
func (l *GameSettlementRollbackLogic) GameSettlementRollback(in *gamepb.GameSettlementRollbackRequest) (*gamepb.GameSettlementRollbackResponse, error) {
	// todo: add your logic here and delete this line

	return &gamepb.GameSettlementRollbackResponse{}, nil
}
