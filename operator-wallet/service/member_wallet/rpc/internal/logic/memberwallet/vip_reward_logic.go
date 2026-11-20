package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type VipRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVipRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VipRewardLogic {
	return &VipRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VIP 奖励入账
func (l *VipRewardLogic) VipReward(in *rewardpb.VipRewardRequest) (*rewardpb.VipRewardResponse, error) {
	// todo: add your logic here and delete this line

	return &rewardpb.VipRewardResponse{}, nil
}
