package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/withdrawpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawalPayoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWithdrawalPayoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalPayoutLogic {
	return &WithdrawalPayoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现出款
func (l *WithdrawalPayoutLogic) WithdrawalPayout(in *withdrawpb.WithdrawalPayoutRequest) (*withdrawpb.WithdrawalPayoutResponse, error) {
	// todo: add your logic here and delete this line

	return &withdrawpb.WithdrawalPayoutResponse{}, nil
}
