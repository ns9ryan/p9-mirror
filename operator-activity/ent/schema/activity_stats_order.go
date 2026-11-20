package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"
)

// ActivityStatsOrder 定义活动订单领取明细统计表结构。
type ActivityStatsOrder struct {
	ent.Schema
}

// Fields 定义活动统计订单表字段
func (ActivityStatsOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_order_id").
			Comment("已计入统计的活动订单 ID，唯一"),

		field.Int64("activity_id").
			Comment("活动 ID"),

		field.Int64("member_id").
			Comment("领取会员 ID"),

		field.Int64("amount").
			Default(0).
			NonNegative().
			Comment("本订单计入统计的领取金额（最小货币单位）"),
	}
}

// Indexes 定义活动统计订单表索引
func (ActivityStatsOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("activity_order_id").
			Unique().
			StorageKey("uk_activity_stats_order_order"),
		index.Fields("activity_id").
			StorageKey("idx_activity_stats_order_activity"),
	}
}

// Mixin 定义活动统计订单表公共字段
func (ActivityStatsOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动统计订单表查询隔离
func (ActivityStatsOrder) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动统计订单表数据库注解
func (ActivityStatsOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动订单领取明细统计表"),
		entsql.Annotation{
			Table: "activity_stats_order",
		},
	}
}
