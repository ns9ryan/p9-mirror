package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/withdrawpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawalFeeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWithdrawalFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalFeeLogic {
	return &WithdrawalFeeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现手续费
func (l *WithdrawalFeeLogic) WithdrawalFee(in *withdrawpb.WithdrawalFeeRequest) (*withdrawpb.WithdrawalFeeResponse, error) {
	// todo: add your logic here and delete this line

	return &withdrawpb.WithdrawalFeeResponse{}, nil
}
