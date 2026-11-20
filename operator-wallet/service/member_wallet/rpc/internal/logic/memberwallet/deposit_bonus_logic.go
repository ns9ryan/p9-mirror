package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DepositBonusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDepositBonusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepositBonusLogic {
	return &DepositBonusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 存款赠送入账
func (l *DepositBonusLogic) DepositBonus(in *depositpb.DepositBonusRequest) (*depositpb.DepositBonusResponse, error) {
	// todo: add your logic here and delete this line

	return &depositpb.DepositBonusResponse{}, nil
}
