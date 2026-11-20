// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package manage

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMemberTaskRewardStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新会员任务奖励状态（只允许从 1 改为 2）
func NewUpdateMemberTaskRewardStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMemberTaskRewardStatusLogic {
	return &UpdateMemberTaskRewardStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMemberTaskRewardStatusLogic) UpdateMemberTaskRewardStatus(req *types.UpdateMemberTaskRewardStatusReq) (resp *types.UpdateMemberTaskRewardStatusResp, err error) {
	// 获取操作员信息
	claims := ctxdata.ClaimsFromCtx(l.ctx)
	operatorCode := claims.OperatorCode

	l.Infof("starting update member task reward status for operator: %s, reward id: %d, new_status: %d", operatorCode, req.Id, req.Status)

	// 构建 RPC 请求
	rpcReq := &taskpb.UpdateMemberTaskRewardStatusReq{
		OperatorCode: operatorCode,
		Id:           req.Id,
		Status:       req.Status,
	}

	// 调用 RPC
	rpcResp, err := l.svcCtx.TaskServiceRpc.MemberTaskRewardUpdateStatus(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("failed to update member task reward status: %v", err)
		return nil, err
	}

	// 转换响应
	reward := rpcResp.GetMemberTaskReward()
	return &types.UpdateMemberTaskRewardStatusResp{
		MemberTaskReward: types.MemberTaskReward{
			Id:                  reward.Id,
			OperatorCode:        reward.OperatorCode,
			MemberId:            reward.MemberId,
			TaskCode:            reward.TaskCode,
			KafkaId:             reward.KafkaId,
			Category:            reward.Category,
			PeriodKey:           reward.PeriodKey,
			PayoutMode:          reward.PayoutMode,
			Amount:              reward.Amount,
			Status:              reward.Status,
			OrderNo:             reward.OrderNo,
			WalletTransactionId: reward.WalletTransactionId,
			DeviceId:            reward.DeviceId,
			WalletId:            reward.WalletId,
			ClaimIp:             reward.ClaimIp,
			PayoutAt:            reward.PayoutAt,
			CreatedAt:           reward.CreatedAt,
			UpdatedAt:           reward.UpdatedAt,
		},
	}, nil
}
