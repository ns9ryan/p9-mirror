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

// ActivityOrderItem 定义活动订单发放明细表结构。
type ActivityOrderItem struct {
	ent.Schema
}

// Fields 定义活动订单项表字段
func (ActivityOrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_order_id").
			Comment("所属活动订单 ID"),

		field.Int64("member_id").
			Comment("会员 ID"),

		field.Int64("activity_id").
			Comment("活动 ID"),

		field.Int("activity_type").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动类型快照，取值见业务常量"),

		field.Int("activity_round").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("活动轮次快照"),

		field.Int64("reward_amount").
			Default(0).
			NonNegative().
			Comment("本项奖励金额"),

		field.String("idempotency_key").
			MaxLen(128).
			NotEmpty().
			Comment("发放幂等键，同厅唯一"),
	}
}

// Indexes 定义活动订单项表索引
func (ActivityOrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "idempotency_key").
			Unique().
			StorageKey("uk_activity_order_item_idempotency"),
		index.Fields("activity_order_id").
			StorageKey("idx_activity_order_item_order"),
	}
}

// Mixin 定义活动订单项表公共字段
func (ActivityOrderItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动订单项表查询隔离
func (ActivityOrderItem) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动订单项表数据库注解
func (ActivityOrderItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动订单发放明细表"),
		entsql.Annotation{
			Table: "activity_order_item",
		},
	}
}
