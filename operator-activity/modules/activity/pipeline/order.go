package pipeline

import (
	"time"

	"oa.98ent.com/p9/common/utils/orderno"
	"oa.98ent.com/p9/operator-activity/constants"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// Builder 奖励订单构建器
type Builder struct {
	orderNoSf *orderno.Snowflake
}

func NewBuilder() *Builder {
	return &Builder{
		// orderNoSf 进程内唯一的订单号生成器。nodeID 为 0 时从内网 IP 计算节点号。
		orderNoSf: orderno.NewSnowflake(0),
	}
}

// Build 根据奖励项构建订单
func (b *Builder) Build(ctx *buildContext, items []types.RewardItem) []*types.ActivityOrder {
	orders := make([]*types.ActivityOrder, 0, len(items))
	loc := time.UTC
	appliedAt := ctx.event.GetEventTime().In(loc)
	if appliedAt.IsZero() {
		appliedAt = time.Now().In(loc)
	}
	createdAt := time.Now().In(loc)

	rewardsTotal := int64(0)
	orderNo := b.orderNoSf.GenerateWithPrefix(constants.ActivityOrderPrefix)
	orderItems := make([]*types.ActivityOrderItem, 0, len(items))
	for _, item := range items {
		rewardsTotal += item.Amount
		orderItems = append(orderItems, &types.ActivityOrderItem{
			OperatorCode:   ctx.activity.OperatorCode,
			MemberID:       ctx.member.ID,
			ActivityID:     ctx.activity.ID,
			ActivityType:   ctx.activity.Type,
			Round:          ctx.round,
			RewardAmount:   item.Amount,
			IdempotencyKey: item.IdempotencyKey,
			CreatedAt:      createdAt,
		})
	}

	orders = append(orders, &types.ActivityOrder{
		OrderNo:       orderNo,
		OperatorCode:  ctx.activity.OperatorCode,
		MemberID:      ctx.member.ID,
		AgentID:       ctx.member.AgentID,
		ActivityID:    ctx.activity.ID,
		ActivityType:  ctx.activity.Type,
		Round:         ctx.round,
		EventType:     ctx.event.GetEventType(),
		RelOrderID:    ctx.event.GetOrderID(),
		RelAmount:     ctx.event.GetAmount(),
		Status:        types.OrderStatusPendingAudit,
		RewardsTotal:  rewardsTotal,
		PayoutMode:    ctx.activity.PayoutMode,
		PayoutTiming:  ctx.activity.PayoutTiming,
		ClientType:    types.ClientType(ctx.event.GetClientType()),
		Remark:        "order build",
		Version:       ctx.activity.Version,
		AppliedAt:     appliedAt,
		SubOrderCount: len(orderItems),
		CreatedAt:     createdAt,
		Items:         orderItems,
	})

	return orders
}

type buildContext struct {
	activity *types.Activity
	event    types.ActivityEvent
	member   *types.MemberInfo
	round    int
}

// BuildFrom 从评估上下文构建
func (b *Builder) BuildFrom(activity *types.Activity, event types.ActivityEvent, member *types.MemberInfo, round int, items []types.RewardItem) []*types.ActivityOrder {
	return b.Build(&buildContext{
		activity: activity,
		event:    event,
		member:   member,
		round:    round,
	}, items)
}
