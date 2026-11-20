package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/withdrawpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawalUnfreezeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWithdrawalUnfreezeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalUnfreezeLogic {
	return &WithdrawalUnfreezeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现解冻退回
func (l *WithdrawalUnfreezeLogic) WithdrawalUnfreeze(in *withdrawpb.WithdrawalUnfreezeRequest) (*withdrawpb.WithdrawalUnfreezeResponse, error) {
	// todo: add your logic here and delete this line

	return &withdrawpb.WithdrawalUnfreezeResponse{}, nil
}
