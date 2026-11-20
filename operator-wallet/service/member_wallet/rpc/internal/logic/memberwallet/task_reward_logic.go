package memberwalletlogic

import (
	"context"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/svc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/rewardpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TaskRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTaskRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskRewardLogic {
	return &TaskRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务奖励入账
func (l *TaskRewardLogic) TaskReward(in *rewardpb.TaskRewardRequest) (*rewardpb.TaskRewardResponse, error) {
	// todo: add your logic here and delete this line

	return &rewardpb.TaskRewardResponse{}, nil
}
