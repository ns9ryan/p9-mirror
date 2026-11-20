package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/withdrawpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawalFreezeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWithdrawalFreezeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawalFreezeLogic {
	return &WithdrawalFreezeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现冻结
func (l *WithdrawalFreezeLogic) WithdrawalFreeze(in *withdrawpb.WithdrawalFreezeRequest) (*withdrawpb.WithdrawalFreezeResponse, error) {
	// todo: add your logic here and delete this line

	return &withdrawpb.WithdrawalFreezeResponse{}, nil
}
