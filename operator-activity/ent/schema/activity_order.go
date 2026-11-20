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

// ActivityOrder 定义活动奖励订单表结构。
type ActivityOrder struct {
	ent.Schema
}

// Fields 定义活动订单表字段
func (ActivityOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(64).
			NotEmpty().
			Comment("订单编号"),

		field.Int64("member_id").
			Comment("领奖/申请会员 ID"),

		field.Int64("agent_id").
			Default(0).
			Comment("会员所属代理 ID，无代理为 0"),

		field.Int64("activity_id").
			Comment("关联活动 ID"),

		field.Int("type").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动类型快照，取值见业务常量"),

		field.Int("round").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("活动轮次快照"),

		field.Int("event_type").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("触发事件类型，取值见业务常量"),

		field.Int64("rel_event_order_id").
			Default(0).
			Comment("关联事件/源订单 ID，无关联为 0"),

		field.Int64("rel_amount").
			Default(0).
			NonNegative().
			Comment("关联业务金额（最小货币单位）"),

		field.Int("status").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("订单状态: 1待审核, 2审核通过, 3审核拒绝, 4待领取, 5已领取, 6已过期, 7待派发, 8已派发"),

		field.Int("audit_status").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("审核状态: 1人工审核, 2风控拦截"),

		field.Time("approved_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("审批时间，未审批为空"),

		field.Int64("approver_uid").
			Default(0).
			Comment("审批人后台用户 ID，没有审批人为 0"),

		field.Int64("rewards_total").
			Default(0).
			NonNegative().
			Comment("奖励总额"),

		field.Int("payout_mode").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("发放方式: 1 手动领取，过期作废, 2 手动领取，过期自动领取, 3 自动派发"),

		field.Int("payout_timing").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("发放时机: 1 实时, 2 次日"),

		field.Time("payout_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("计划/实际发放时间，未定为空"),

		field.Time("claim_expire_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("领取截止时间，过期不可领，未限制为空"),

		field.String("remark").
			MaxLen(512).
			Default("").
			Comment("订单备注"),

		field.Int("version").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("订单版本号，乐观锁"),

		field.Int("sub_order_count").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("子订单/发放项数量"),

		field.Time("applied_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("用户申请/触发时间"),

		field.Int("channel_source").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("渠道来源: 1 PC网页, 2 移动网页, 3 Android-APP, 4 IOS-APP"),
	}
}

// Indexes 定义活动订单表索引
func (ActivityOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "order_no").
			Unique().
			StorageKey("uk_activity_order_operator_order_no"),
		index.Fields("operator_code", "activity_id").
			StorageKey("idx_activity_order_operator_activity_id"),
		index.Fields("operator_code", "member_id").
			StorageKey("idx_activity_order_operator_member"),
	}
}

// Mixin 定义活动订单表公共字段
func (ActivityOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动订单表查询隔离
func (ActivityOrder) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动订单表数据库注解
func (ActivityOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动奖励订单表"),
		entsql.Annotation{
			Table: "activity_order",
		},
	}
}
