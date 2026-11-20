package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type VIPRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVIPRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VIPRewardLogic {
	return &VIPRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VIP 奖励入账
func (l *VIPRewardLogic) VIPReward(in *rewardpb.VIPRewardRequest) (*rewardpb.VIPRewardResponse, error) {
	// todo: add your logic here and delete this line

	return &rewardpb.VIPRewardResponse{}, nil
}
