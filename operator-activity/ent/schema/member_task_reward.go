package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
)

// MemberTaskReward 定义奖励订单表结构
type MemberTaskReward struct {
	ent.Schema
}

// Fields 定义奖励订单字段
func (MemberTaskReward) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("member_id").
			Comment("会员 ID"),

		field.String("task_code").
			MaxLen(64).
			NotEmpty().
			Comment("任务编码"),

		field.String("kafka_id").
			MaxLen(128).
			Comment("Kafka ID"),

		field.String("category").
			MaxLen(16).
			NotEmpty().
			Comment("任务分类：ONCE|DAILY|WEEKLY|MONTHLY"),

		field.String("period_key").
			MaxLen(32).
			NotEmpty().
			Comment("周期键：ALL|20261103|2026-W45|202611"),

		field.String("payout_mode").
			MaxLen(16).
			NotEmpty().
			Comment("派发方式：MANUAL|AUTO"),

		field.Int64("amount").
			NonNegative().
			Comment("金额（最小货币单位）"),

		field.Int("status").
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("订单状态：1=待领取|2=待派发|11=任务已拒绝|12=派发被拒绝|99=已派发"),

		field.String("order_no").
			MaxLen(64).
			Unique().
			NotEmpty().
			Comment("订单号"),

		field.String("wallet_transaction_id").
			MaxLen(128).
			Optional().
			Comment("钱包交易 ID"),

		field.String("device_id").
			MaxLen(128).
			Optional().
			Comment("领取设备 ID"),

		field.String("wallet_id").
			MaxLen(128).
			Optional().
			Comment("领取钱包 ID"),

		field.String("claim_ip").
			MaxLen(45).
			Optional().
			Comment("领取 IP"),

		field.Time("payout_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("派发时间/领取时间/钱包实际变更时间"),
	}
}

// Indexes 定义奖励订单表索引
func (MemberTaskReward) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "member_id", "task_code", "period_key").
			Unique().
			StorageKey("idx_reward_unique"),
		index.Fields("status").
			StorageKey("idx_reward_status"),
		index.Fields("order_no").
			StorageKey("idx_reward_order_no"),
	}
}

// Mixin 定义奖励订单表公共字段
func (MemberTaskReward) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// // Interceptors 定义奖励订单表查询隔离
// func (MemberTaskReward) Interceptors() []ent.Interceptor {
// 	return []ent.Interceptor{
// 		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
// 			entmixin.FilterOperatorCode(ctx, q)
// 			return nil
// 		}),
// 	}
// }

// Annotations 定义奖励订单表数据库注解
func (MemberTaskReward) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("任务奖励订单表"),
		entsql.Annotation{
			Table: "member_task_reward",
		},
	}
}
