package schema

import (
	"context"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ActivityWheelChanceGrant 活动幸运转盘抽奖机会发放记录
type ActivityWheelChanceGrant struct{ ent.Schema }

func (ActivityWheelChanceGrant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").Comment("活动ID"),
		field.Int64("member_id").Comment("会员ID"),
		field.Int("round").Default(0).Comment("周期轮次"),
		field.Int64("schema_id").Default(0).Comment("SchemaID"),
		field.Int("event_type").Default(0).Comment("事件类型"),
		field.Int64("order_id").Default(0).Comment("订单ID"),
		field.Int("chances").Default(0).Comment("抽奖机会"),
		field.String("idempotency_key").MaxLen(128).Default("").Comment("幂等键"),
	}
}

func (ActivityWheelChanceGrant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("idempotency_key").Unique().StorageKey("uk_wheel_chance_grant"),
	}
}

func (ActivityWheelChanceGrant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动统计表查询隔离
func (ActivityWheelChanceGrant) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动统计表数据库注解
func (ActivityWheelChanceGrant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动幸运转盘抽奖机会发放记录表"),
		entsql.Annotation{
			Table: "activity_wheel_chance_grant",
		},
	}
}
