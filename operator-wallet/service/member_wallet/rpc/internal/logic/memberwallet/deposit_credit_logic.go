package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DepositCreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDepositCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepositCreditLogic {
	return &DepositCreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 存款本金入账
func (l *DepositCreditLogic) DepositCredit(in *depositpb.DepositCreditRequest) (*depositpb.DepositCreditResponse, error) {
	// todo: add your logic here and delete this line

	return &depositpb.DepositCreditResponse{}, nil
}
