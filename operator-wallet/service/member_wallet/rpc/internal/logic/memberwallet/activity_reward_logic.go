package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivityRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewActivityRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivityRewardLogic {
	return &ActivityRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 活动奖励入账
func (l *ActivityRewardLogic) ActivityReward(in *rewardpb.ActivityRewardRequest) (*rewardpb.ActivityRewardResponse, error) {
	// todo: add your logic here and delete this line

	return &rewardpb.ActivityRewardResponse{}, nil
}
