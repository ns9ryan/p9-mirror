package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/gamepb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GameTipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGameTipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GameTipLogic {
	return &GameTipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 游戏消费打赏
func (l *GameTipLogic) GameTip(in *gamepb.GameTipRequest) (*gamepb.GameTipResponse, error) {
	// todo: add your logic here and delete this line

	return &gamepb.GameTipResponse{}, nil
}
