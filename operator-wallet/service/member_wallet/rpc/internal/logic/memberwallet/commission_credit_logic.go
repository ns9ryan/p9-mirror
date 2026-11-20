package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommissionCreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCommissionCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommissionCreditLogic {
	return &CommissionCreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 全民代理佣金入账
func (l *CommissionCreditLogic) CommissionCredit(in *rewardpb.CommissionCreditRequest) (*rewardpb.CommissionCreditResponse, error) {
	// todo: add your logic here and delete this line

	return &rewardpb.CommissionCreditResponse{}, nil
}
