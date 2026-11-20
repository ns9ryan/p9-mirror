package promotion

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/audit"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/ctxdata"
)

// Engine 活动奖励编排引擎：加载活动 → 使用插件流水线评估 → 发奖
type Engine struct {
	repo    *repository.Repository
	builder *pipeline.Builder
	locker  UserPromoLocker // 可选；nil 时不加锁
}

// UserPromoLocker 同一用户同一活动的发奖串行化（通常为 Redis 分布式锁）
type UserPromoLocker interface {
	Lock(ctx context.Context, promoID, userID int64) (unlock func(), err error)
}

func NewEngine(repo *repository.Repository, locker UserPromoLocker) *Engine {
	return &Engine{
		repo:    repo,
		builder: pipeline.NewBuilder(),
		locker:  locker,
	}
}

// ProcessEvent 处理单个事件：加载活动 → 逐条评估 → 发奖。
// pipelines 由插件 Service 提供，定义本业务的指标/条件/奖励/派发策略。
// memberInfo 由上层透传时可避免重复查库；为 nil 时引擎自行 LoadUser。
// 任一活动 processOne 失败则立即返回（经 perr 分类），由上层决定重试或跳过。
func (e *Engine) ProcessEvent(
	ctx context.Context,
	evt types.ActivityEvent,
	activityType types.ActivityType,
	memberInfo *types.MemberInfo,
	pipelines plugin.PipelineFactory,
) error {
	if evt != nil && ctxdata.OperatorCodeFromCtx(ctx) == "" && evt.GetOperatorCode() != "" {
		ctx = ctxdata.WithOperatorCode(ctx, evt.GetOperatorCode())
	}
	activities, err := e.repo.LoadOngoingActivities(ctx, activityType)
	if err != nil {
		return perr.Retryable(fmt.Errorf("load activities: %w", err))
	}
	if len(activities) == 0 {
		logx.Infof("[engine] no ongoing activities operator_code=%s type=%d", evt.GetOperatorCode(), activityType)
		return nil
	}
	// 遍历活动，逐个处理
	for _, activity := range activities {
		// 构建流水线
		pipe := pipelines.BuildPipeline(activity)
		// 处理活动
		if err := e.processOne(ctx, activity, evt, memberInfo, pipe); err != nil {
			err = perr.Classify(err)
			logx.Errorf("[engine] process activity=%d err=%v retryable=%v",
				activity.ID, err, perr.IsRetryable(err))
			return err
		}
	}
	return nil
}

func (e *Engine) processOne(
	ctx context.Context,
	activity *types.Activity,
	evt types.ActivityEvent,
	memberInfo *types.MemberInfo,
	pipe plugin.ActivityPipeline,
) error {
	if e.locker != nil {
		unlock, err := e.locker.Lock(ctx, activity.ID, memberInfo.ID)
		if err != nil {
			return err
		}
		defer unlock()
	}
	// 创建步骤调试日志
	stepLog := pipeline.NewStepLogger(activity.ID, memberInfo.ID)
	// 创建条件、奖励、派发上下文
	evalCtx := &pipeline.EvalContext{
		Context:  ctx,
		Activity: activity,
		Event:    evt,
		Member:   memberInfo,
		StepLog:  stepLog,
	}
	defer stepLog.Flush()

	// 开始处理活动
	stepLog.Stepf("process", "start", "operator_code=%s", evt.GetOperatorCode())

	switch pipe.Conditions().Run(evalCtx) {
	case pipeline.ResultError:
		stepLog.Step("condition_chain", "error", "condition evaluate error")
		return perr.Retryablef("condition chain error activity=%d", activity.ID)
	case pipeline.ResultPass:
		stepLog.Step("condition_chain", "pass", "all conditions passed")
	default:
		stepLog.Step("condition_chain", "skip", "not all conditions passed")
		return nil
	}
	// 计算奖励
	rewards, err := pipe.Rewards().Calculate(evalCtx)
	if err != nil {
		stepLog.Stepf("reward", "error", "err=%v", err)
		return perr.Permanent(fmt.Errorf("calculate reward: %w", err))
	}
	if len(rewards) == 0 {
		stepLog.Step("reward", "skip", "no reward matched")
		return nil
	}
	stepLog.Stepf("reward", "pass", "matched=%d first_amount=%d key=%s",
		len(rewards), rewards[0].Amount, rewards[0].IdempotencyKey)

	var filtered []types.RewardItem
	for _, r := range rewards {
		if r.Amount <= 0 {
			stepLog.Stepf("reward", "skip", "amount<=0 schema_id=%d", r.SchemaID)
			continue
		}
		// 检查幂等性
		exists, err := e.repo.IdempKeyExists(ctx, r.IdempotencyKey)
		if err != nil {
			stepLog.Stepf("idempotency", "error", "key=%s err=%v", r.IdempotencyKey, err)
			return perr.Retryable(fmt.Errorf("idempotency check: %w", err))
		}
		stepLog.Stepf("idempotency", "pass", "key=%s exists=%v", r.IdempotencyKey, exists)
		if !exists {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		stepLog.Step("idempotency", "skip", "all rewards already exist")
		return nil
	}
	// 奖励订单审核策略
	auditStrat := pipe.Audit()
	// 派发奖励策略
	payoutStrategy := pipe.Payout()
	// 构建奖励订单
	orders := e.builder.BuildFrom(activity, evt, memberInfo, evalCtx.Round, filtered)
	for _, o := range orders {
		switch auditStrat.Decide(o) {
		case audit.DecisionAutoPass:
			o.Status = types.OrderStatusApproved
			o.ApprovedAt = o.CreatedAt // 自动审核通过时，应用时间与创建时间一致
			if o.Remark == "" {
				o.Remark = "auto audit passed"
			}
			// 应用派发策略
			payoutStrategy.Apply(ctx, o)
			stepLog.Stepf("audit", "pass",
				"order_no=%s decision=auto_pass amount=%d status=%d",
				o.OrderNo, o.RewardsTotal, o.Status)
			stepLog.Stepf("payout_apply", "pass",
				"order_no=%s status=%d payout_at=%s claim_exp=%s mode=%d",
				o.OrderNo, o.Status,
				o.PayoutAt.Format("2006-01-02 15:04:05"),
				o.ClaimExp.Format("2006-01-02 15:04:05"),
				o.PayoutMode)
		default:
			o.Status = types.OrderStatusPendingAudit
			o.AuditStatus = types.AuditStatusManual
			stepLog.Stepf("audit", "pending",
				"order_no=%s decision=manual amount=%d", o.OrderNo, o.RewardsTotal)
		}
	}
	// 创建活动奖励订单
	if err := e.repo.SaveOrders(ctx, orders, activity, memberInfo); err != nil {
		stepLog.Stepf("save_orders", "error", "count=%d err=%v", len(orders), err)
		return perr.Classify(fmt.Errorf("save orders: %w", err))
	}
	stepLog.Stepf("save_orders", "pass", "count=%d", len(orders))
	// 派发奖励
	for _, o := range orders {
		// 跳过待审核订单，后续人工审核通过后会重新调度派发
		if o.Status == types.OrderStatusPendingAudit {
			continue
		}
		// 调度派发
		if err := payoutStrategy.Schedule(ctx, o); err != nil {
			stepLog.Stepf("payout_schedule", "error", "order_id=%d err=%v", o.ID, err)
			return perr.Retryable(fmt.Errorf("payout schedule: %w", err))
		}
		stepLog.Stepf("payout_schedule", "pass", "order_id=%d amount=%d", o.ID, o.RewardsTotal)
	}
	// 发布奖励订单创建事件
	if err := e.repo.PublishOrderCreated(ctx, orders); err != nil {
		stepLog.Stepf("publish_created", "error", "count=%d err=%v", len(orders), err)
		return perr.Retryable(fmt.Errorf("publish order created: %w", err))
	}
	stepLog.Stepf("publish_created", "pass", "count=%d", len(orders))
	stepLog.Stepf("engine", "pass", "created %d orders", len(orders))
	return nil
}

// // ApproveOrder 人工审核通过：按插件流水线 Apply 派发字段 → 落库 → 启动领取/派发工作流
// func (e *Engine) ApproveOrder(ctx context.Context, orderID int64, factories plugin.PipelineFactory) error {
// 	if factories == nil {
// 		return perr.Permanentf("approve order: pipeline factory required")
// 	}
// 	order, err := e.repo.GetOrder(ctx, orderID)
// 	if err != nil {
// 		return perr.Classify(fmt.Errorf("get order: %w", err))
// 	}
// 	if order.Status != types.OrderStatusPendingAudit {
// 		return perr.Permanentf("order %d status=%d not pending audit", orderID, order.Status)
// 	}

// 	activity, err := e.repo.GetActivity(ctx, order.ActivityID)
// 	if err != nil {
// 		return perr.Classify(fmt.Errorf("get activity: %w", err))
// 	}
// 	pipe := factories.BuildPipeline(activity)

// 	order.Status = types.OrderStatusApproved
// 	order.Remark = "人工审核通过"
// 	payoutStrategy := pipe.Payout()
// 	payoutStrategy.Apply(ctx, order)

// 	if err := e.repo.UpdateOrderPayout(ctx, order); err != nil {
// 		return perr.Classify(fmt.Errorf("update order after approve: %w", err))
// 	}
// 	if err := payoutStrategy.Schedule(ctx, order); err != nil {
// 		return perr.Retryable(fmt.Errorf("payout schedule after approve: %w", err))
// 	}
// 	logx.Infof("[engine] order=%d approved status=%d payout_at=%v", order.ID, order.Status, order.PayoutAt)
// 	return nil
// }

// // RejectOrder 人工审核驳回
// func (e *Engine) RejectOrder(ctx context.Context, orderID int64, remark string) error {
// 	order, err := e.repo.GetOrder(ctx, orderID)
// 	if err != nil {
// 		return perr.Classify(fmt.Errorf("get order: %w", err))
// 	}
// 	if order.Status != types.OrderStatusPendingAudit {
// 		return perr.Permanentf("order %d status=%d not pending audit", orderID, order.Status)
// 	}
// 	if remark == "" {
// 		remark = "manual audit rejected"
// 	}
// 	if err := e.repo.UpdateOrderStatus(ctx, orderID, types.OrderStatusRejected, remark); err != nil {
// 		return perr.Classify(fmt.Errorf("reject order: %w", err))
// 	}
// 	logx.Infof("[engine] order=%d rejected", orderID)
// 	return nil
// }
