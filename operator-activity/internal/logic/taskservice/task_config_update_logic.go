package taskservicelogic

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/ent/taskconfig"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/taskpb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type TaskConfigUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTaskConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskConfigUpdateLogic {
	return &TaskConfigUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// group: task_config 更新任务配置（仅允许修改 timezone、status、completion_amount、reward_amount、payout_mode、sort）
func (l *TaskConfigUpdateLogic) TaskConfigUpdate(in *taskpb.UpdateTaskConfigReq) (*taskpb.UpdateTaskConfigResp, error) {
	l.Infof("starting task config update for operator: %s, task id: %d", in.OperatorCode, in.Id)
	// 验证操作员权限
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 验证操作员代码一致
	if in.OperatorCode != claims.OperatorCode {
		l.Infof("starting task config update for operator: %s, operator code in claims: %v", in.OperatorCode, claims.OperatorCode)
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	// 获取任务配置
	config, err := l.svcCtx.DB.TaskConfig.Query().
		Where(
			taskconfig.ID(in.Id),
			taskconfig.OperatorCode(in.OperatorCode),
		).
		Only(l.ctx)
	if err != nil {
		l.Errorf("failed to query task config: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 构建更新条件
	updateBuilder := l.svcCtx.DB.TaskConfig.UpdateOne(config)

	// 仅允许修改 timezone、status、completion_amount、reward_amount、payout_mode、sort
	if in.Timezone != nil {
		updateBuilder = updateBuilder.SetTimezone(*in.Timezone)
	}
	if in.Status != nil {
		updateBuilder = updateBuilder.SetStatus(int(*in.Status))
	}
	if in.CompletionAmount != nil {
		updateBuilder = updateBuilder.SetCompletionAmount(int(*in.CompletionAmount))
	}
	if in.RewardAmount != nil {
		updateBuilder = updateBuilder.SetRewardAmount(*in.RewardAmount)
	}
	if in.PayoutMode != nil {
		updateBuilder = updateBuilder.SetPayoutMode(*in.PayoutMode)
	}
	if in.Sort != nil {
		updateBuilder = updateBuilder.SetSort(*in.Sort)
	}

	// 执行更新
	updated, err := updateBuilder.Save(l.ctx)
	if err != nil {
		l.Errorf("failed to update task config: %v", err)
		return nil, xerr.RpcErr(err)
	}

	// 更新数据库成功后，清除 Redis 缓存（本实例立即清缓存）
	if l.svcCtx.RedisQueue != nil {
		if cacheErr := l.svcCtx.RedisQueue.InvalidateTaskConfigCache(l.ctx, updated.OperatorCode, updated.TaskCode); cacheErr != nil {
			// 缓存清除失败不影响业务，仅记录日志
			l.Errorf("⚠️ [缓存更新失败] task_config 缓存清除失败: operatorCode=%s, taskCode=%s, err=%v",
				updated.OperatorCode, updated.TaskCode, cacheErr)
		} else {
			l.Infof("✓ [缓存更新成功] task_config 配置已更新并清除缓存: operatorCode=%s, taskCode=%s",
				updated.OperatorCode, updated.TaskCode)
		}
	}

	return &taskpb.UpdateTaskConfigResp{
		TaskConfig: &taskpb.TaskConfig{
			Id:               updated.ID,
			OperatorCode:     updated.OperatorCode,
			TaskCode:         updated.TaskCode,
			Category:         updated.Category,
			Name:             updated.Name,
			NameI18N:         updated.NameI18n,
			Description:      updated.Description,
			DescriptionI18N:  updated.DescriptionI18n,
			Timezone:         updated.Timezone,
			Status:           int32(updated.Status),
			CompletionAmount: int32(updated.CompletionAmount),
			RewardAmount:     updated.RewardAmount,
			PayoutMode:       updated.PayoutMode,
			Sort:             updated.Sort,
			CreatedAt:        updated.CreatedAt.UnixMilli(),
			UpdatedAt:        updated.UpdatedAt.UnixMilli(),
		},
	}, nil
}
