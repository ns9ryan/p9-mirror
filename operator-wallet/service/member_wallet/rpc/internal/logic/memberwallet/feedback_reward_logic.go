package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type FeedbackRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFeedbackRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedbackRewardLogic {
	return &FeedbackRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 有奖反馈奖励入账
func (l *FeedbackRewardLogic) FeedbackReward(in *rewardpb.FeedbackRewardRequest) (*rewardpb.FeedbackRewardResponse, error) {
	// todo: add your logic here and delete this line

	return &rewardpb.FeedbackRewardResponse{}, nil
}
