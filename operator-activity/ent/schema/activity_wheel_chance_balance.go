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

// ActivityWheelChanceBalance 活动幸运转盘抽奖机会余额表
type ActivityWheelChanceBalance struct{ ent.Schema }

func (ActivityWheelChanceBalance) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").Comment("活动ID"),
		field.Int64("member_id").Comment("会员ID"),
		field.Int("granted").Default(0).Comment("发放数量"),
		field.Int("used").Default(0).Comment("使用数量"),
	}
}

func (ActivityWheelChanceBalance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("activity_id", "member_id").Unique().StorageKey("uk_wheel_chance_balance"),
	}
}

func (ActivityWheelChanceBalance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动统计表查询隔离
func (ActivityWheelChanceBalance) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动统计表数据库注解
func (ActivityWheelChanceBalance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动幸运转盘抽奖机会余额表"),
		entsql.Annotation{
			Table: "activity_wheel_chance_balance",
		},
	}
}
