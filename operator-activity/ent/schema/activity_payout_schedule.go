package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"
)

// ActivityPayoutSchedule 定义活动定时发放计划表结构。
type ActivityPayoutSchedule struct {
	ent.Schema
}

// Fields 定义活动发放计划表字段
func (ActivityPayoutSchedule) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").
			Comment("所属活动 ID"),

		field.Int64("activity_order_id").
			Comment("所属活动订单 ID"),

		field.Int64("member_id").
			Comment("会员 ID"),

		field.Int64("rewards_amount").
			Default(0).
			NonNegative().
			Comment("奖励金额"),

		field.Time("payout_at").
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("计划发放时间"),

		field.Int("status").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("发放计划状态，取值见业务常量"),
	}
}

// Indexes 定义活动发放计划表索引
func (ActivityPayoutSchedule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "payout_at").
			StorageKey("idx_activity_payout_schedule_status_payout_at"),
		index.Fields("activity_id").
			StorageKey("idx_activity_payout_schedule_activity"),
	}
}

// Mixin 定义活动发放计划表公共字段
func (ActivityPayoutSchedule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动发放计划表查询隔离
func (ActivityPayoutSchedule) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动发放计划表数据库注解
func (ActivityPayoutSchedule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动定时发放计划表"),
		entsql.Annotation{
			Table: "activity_payout_schedule",
		},
	}
}
