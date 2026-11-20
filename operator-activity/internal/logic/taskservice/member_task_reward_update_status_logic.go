package taskservicelogic

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/membertaskreward"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"
	"oa.98ent.com/p9/operator-activity/task/constants"
	tasktype "oa.98ent.com/p9/operator-activity/task/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MemberTaskRewardUpdateStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMemberTaskRewardUpdateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MemberTaskRewardUpdateStatusLogic {
	return &MemberTaskRewardUpdateStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: member_task_reward 更新会员任务奖励状态（只允许从 1 改为 2）
func (l *MemberTaskRewardUpdateStatusLogic) MemberTaskRewardUpdateStatus(in *taskpb.UpdateMemberTaskRewardStatusReq) (*taskpb.UpdateMemberTaskRewardStatusResp, error) {
	l.Infof("starting member task reward update status for operator: %s, reward id: %d, new_status: %d", in.OperatorCode, in.Id, in.Status)

	// 验证操作员权限
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 验证操作员代码一致
	if in.OperatorCode != claims.OperatorCode {
		l.Infof("operator code mismatch: request operator: %s, claims operator: %v", in.OperatorCode, claims.OperatorCode)
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 验证新状态只能是待派发
	if in.Status != constants.TaskStatusPendingDispatch {
		l.Infof("invalid new status: %d, only %d (待派发) is allowed", in.Status, constants.TaskStatusPendingDispatch)
		return nil, xerr.RpcErr(xerr.BadRequest("only_allow_status_2"))
	}

	// 获取奖励订单
	reward, err := l.svcCtx.DB.MemberTaskReward.Query().
		Where(
			membertaskreward.ID(in.Id),
			membertaskreward.OperatorCode(in.OperatorCode),
		).
		Only(l.ctx)
	if err != nil {
		l.Errorf("failed to query member task reward: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 验证当前状态必须是待领取
	if reward.Status != constants.TaskStatusPendingClaim {
		l.Infof("invalid current status: %d, only %d (待领取) can be updated, reward id: %d", reward.Status, constants.TaskStatusPendingClaim, in.Id)
		return nil, xerr.RpcErr(xerr.BadRequest("only_allow_from_status_1"))
	}

	// 更新状态
	updated, err := l.svcCtx.DB.MemberTaskReward.UpdateOne(reward).
		SetStatus(int(in.Status)).
		Save(l.ctx)
	if err != nil {
		l.Errorf("failed to update member task reward status: %v", err)
		return nil, xerr.RpcErr(err)
	}
	l.DispatchPayoutEventToBatch(updated)
	l.Infof("成功更新会员任务奖励状态: %d, old status: %d, new status: %d", updated.ID, reward.Status, updated.Status)

	payoutAt := int64(0)
	if updated.PayoutAt != nil {
		payoutAt = updated.PayoutAt.UnixMilli()
	}

	return &taskpb.UpdateMemberTaskRewardStatusResp{
		MemberTaskReward: &taskpb.MemberTaskReward{
			Id:                  updated.ID,
			OperatorCode:        updated.OperatorCode,
			MemberId:            updated.MemberID,
			TaskCode:            updated.TaskCode,
			KafkaId:             updated.KafkaID,
			Category:            updated.Category,
			PeriodKey:           updated.PeriodKey,
			PayoutMode:          updated.PayoutMode,
			Amount:              updated.Amount,
			Status:              int32(updated.Status),
			OrderNo:             updated.OrderNo,
			WalletTransactionId: updated.WalletTransactionID,
			DeviceId:            updated.DeviceID,
			WalletId:            updated.WalletID,
			ClaimIp:             updated.ClaimIP,
			PayoutAt:            payoutAt,
			CreatedAt:           updated.CreatedAt.UnixMilli(),
			UpdatedAt:           updated.UpdatedAt.UnixMilli(),
		},
	}, nil
}

func (l *MemberTaskRewardUpdateStatusLogic) DispatchPayoutEventToBatch(taskReward *ent.MemberTaskReward) error {
	if l.svcCtx.RedisQueue == nil {
		l.Errorf("RedisQueue is not initialized")
		return fmt.Errorf("RedisQueue is not initialized")
	}
	// 创建派发事件
	payoutEvent := &tasktype.Event{
		OperatorCode: taskReward.OperatorCode,
		MemberID:     taskReward.MemberID,
		TaskCode:     taskReward.TaskCode,
		Amount:       taskReward.Amount,
		OrderNo:      taskReward.OrderNo,   // 派发时需要的订单号
		PeriodKey:    taskReward.PeriodKey, // 继承原事件的周期键
		KafkaID:      taskReward.KafkaID,   // 追踪源消息
		OccurredAt:   time.Now(),
		RetryCount:   0,
		MaxRetry:     constants.MaxPayoutRetries,
	}

	// 写入派发队列
	if err := l.svcCtx.RedisQueue.DispatchPayoutEventToBatch(l.ctx, payoutEvent); err != nil {
		l.Errorf("写入派发事件到redis失败: order_no=%s, err=%v", taskReward.OrderNo, err)
		return fmt.Errorf("写入派发事件失败: %w", err)
	}
	l.Infof("派发事件已写入redis: order_no=%s", taskReward.OrderNo)
	return nil
}
