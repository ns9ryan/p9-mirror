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

// ActivityStats 定义活动领取汇总统计表结构。
// 金额按最小货币单位保存。
type ActivityStats struct {
	ent.Schema
}

// Fields 定义活动统计表字段
func (ActivityStats) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").
			Comment("被统计的活动 ID"),

		field.Int64("claim_member_count").
			Default(0).
			NonNegative().
			Comment("累计成功领取会员数（去重）"),

		field.Int64("claimed_order_count").
			Default(0).
			NonNegative().
			Comment("累计已领取订单数"),

		field.Int64("claimed_amount").
			Default(0).
			NonNegative().
			Comment("已领取金额（最小货币单位）"),

		field.Int64("unclaimed_amount").
			Default(0).
			NonNegative().
			Comment("未领取金额（最小货币单位）"),

		field.Int64("expired_amount").
			Default(0).
			NonNegative().
			Comment("已过期金额（最小货币单位）"),

		field.Int64("rejected_amount").
			Default(0).
			NonNegative().
			Comment("已拒绝金额（最小货币单位）"),

		field.Int64("pending_audit_amount").
			Default(0).
			NonNegative().
			Comment("待审核金额（最小货币单位）"),

		field.Int64("risk_blocked_amount").
			Default(0).
			NonNegative().
			Comment("风控拦截金额（最小货币单位）"),
	}
}

// Indexes 定义活动统计表索引
func (ActivityStats) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("activity_id").
			Unique().
			StorageKey("uk_activity_stats_activity"),
	}
}

// Mixin 定义活动统计表公共字段
func (ActivityStats) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动统计表查询隔离
func (ActivityStats) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动统计表数据库注解
func (ActivityStats) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动领取汇总统计表"),
		entsql.Annotation{
			Table: "activity_stats",
		},
	}
}
