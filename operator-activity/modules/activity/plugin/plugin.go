package plugin

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

const (
	NameDeposit    = "deposit"
	NameBetting    = "betting"
	NameReferral   = "referral"
	NameLuckyWheel = "luckywheel"
)

// EventProcessor 活动引擎抽象，供插件回调发奖（避免插件依赖 promotion 包造成循环引用）
type EventProcessor interface {
	// ProcessEvent 处理事件；user 可为 nil 时引擎自行 LoadUser；pipelines 由插件 Service 提供
	ProcessEvent(
		ctx context.Context,
		evt types.ActivityEvent,
		activityType types.ActivityType,
		memberInfo *types.MemberInfo,
		pipelines PipelineFactory,
	) error
}

// Plugin 活动插件。调度只投递 Kafka 消息，插件在内部解析成自己的事件。
type Plugin interface {
	Name() string
	ActivityType() types.ActivityType
	Dispatch(ctx context.Context, payload types.Payload) error
}

// DispatchFunc 插件消息处理函数
type DispatchFunc func(ctx context.Context, payload types.Payload) error
