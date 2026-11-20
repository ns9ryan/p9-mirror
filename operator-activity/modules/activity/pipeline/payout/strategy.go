package payout

import (
	"context"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// Strategy 派发策略
type Strategy interface {
	// Apply 仅填充订单派发相关字段（PayoutAt / ClaimExp / Status），无副作用
	Apply(ctx context.Context, order *types.ActivityOrder)
	// Schedule 执行派发副作用（Temporal / 即时发布），依赖 Apply 已填充的字段
	Schedule(ctx context.Context, order *types.ActivityOrder) error
}

// TimingStrategy 派发时间策略
type TimingStrategy interface {
	CalcExecuteTime(ctx context.Context, order *types.ActivityOrder) time.Time
}

// ScheduleStore 延迟派发存储（对账/补偿，主路径走 Temporal）
type ScheduleStore interface {
	SaveSchedule(ctx context.Context, order *types.ActivityOrder) error
}

// PayoutPublisher 派发消息发布
type PayoutPublisher interface {
	PublishPayout(ctx context.Context, order *types.ActivityOrder) error
}

// RealtimeTiming 实时派发
type RealtimeTiming struct{}

func (t *RealtimeTiming) CalcExecuteTime(ctx context.Context, o *types.ActivityOrder) time.Time {
	return o.CreatedAt
}

// NextDayTiming 次日派发
type NextDayTiming struct{}

func (t *NextDayTiming) CalcExecuteTime(ctx context.Context, order *types.ActivityOrder) time.Time {
	loc := time.Local
	created := order.CreatedAt.In(loc)
	next := time.Date(created.Year(), created.Month(), created.Day()+1, 0, 0, 0, 0, loc)
	next = created.Add(120 * time.Second) // 延迟120秒

	return next
}

// AutoPayout 自动派发
type AutoPayout struct {
	Timing    TimingStrategy
	Store     ScheduleStore
	Publisher PayoutPublisher
	Workflows WorkflowStarter
	Timezone  string
}

func (p *AutoPayout) Apply(ctx context.Context, order *types.ActivityOrder) {
	order.PayoutAt = p.Timing.CalcExecuteTime(ctx, order)
	if order.Status == types.OrderStatusApproved && order.PayoutAt.Before(time.Now()) {
		// 自动派发：过审后并且实时到账
		order.Status = types.OrderStatusPendingPayout
	}
}

func (p *AutoPayout) Schedule(ctx context.Context, order *types.ActivityOrder) error {
	if order.PayoutAt.IsZero() {
		p.Apply(ctx, order)
	}

	// 未来到账：写 schedule 对账 + 启 Temporal Timer
	if order.PayoutAt.After(time.Now()) {
		if p.Store != nil {
			if err := p.Store.SaveSchedule(ctx, order); err != nil {
				return err
			}
		}
		if p.Workflows != nil {
			return p.Workflows.StartOrderLifecycle(ctx, order)
		}
		logx.Infof("[payout] order=%d scheduled at %v (no workflow starter)", order.ID, order.PayoutAt)
		return nil
	}

	// 实时：直接发布，同时启短生命周期 WF（可选，便于统一观测）
	if p.Workflows != nil {
		return p.Workflows.StartOrderLifecycle(ctx, order)
	}
	if p.Publisher != nil {
		return p.Publisher.PublishPayout(ctx, order)
	}
	logx.Infof("[payout] order=%d auto payout immediate amount=%d", order.ID, order.RewardsTotal)
	return nil
}

// ExpireAction 过期行为
type ExpireAction int

const (
	ExpireVoid ExpireAction = 1
	ExpireAuto ExpireAction = 2
)

// ManualClaimPayout 手动领取
type ManualClaimPayout struct {
	Timing       TimingStrategy
	Store        ScheduleStore
	Publisher    PayoutPublisher
	Workflows    WorkflowStarter
	ExpireSec    int64
	ExpireAction ExpireAction
}

func (p *ManualClaimPayout) Apply(ctx context.Context, order *types.ActivityOrder) {
	order.PayoutAt = p.Timing.CalcExecuteTime(ctx, order)
	if order.Status == types.OrderStatusApproved && order.PayoutAt.Before(time.Now()) {
		// 手动领取：过审后并且实时到账
		order.Status = types.OrderStatusPendingClaim
	}
	// 手动领取：过审后并且领取状态，设置过期时间
	if p.ExpireSec > 0 && order.Status == types.OrderStatusPendingClaim {
		order.ClaimExp = order.PayoutAt.Add(time.Duration(p.ExpireSec) * time.Second)
	}
}

func (p *ManualClaimPayout) Schedule(ctx context.Context, order *types.ActivityOrder) error {
	if order.PayoutAt.IsZero() {
		p.Apply(ctx, order)
	}
	if p.Workflows != nil {
		return p.Workflows.StartOrderLifecycle(ctx, order)
	}
	logx.Infof("[payout] order=%d pending claim expire_action=%d (no workflow starter)", order.ID, p.ExpireAction)
	return nil
}
