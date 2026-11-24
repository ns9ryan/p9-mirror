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
	"oa.98ent.com/p9/operator-crm/ent/intercept"
	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// Turnover 定义模拟营业额表结构。
type Turnover struct {
	ent.Schema
}

// Fields 定义模拟营业额表字段
func (Turnover) Fields() []ent.Field {
	return []ent.Field{
		// 会员ID
		field.Int64("member_id").
			Positive().
			Comment("会员ID"),

		// 订单号
		field.String("order_no").
			Optional().
			MaxLen(64).
			Comment("订单号"),

		// 消费数据
		field.Int64("consume_amount").
			Default(0).
			Comment("消费金额（分）"),

		field.String("consume_currency").
			Optional().
			MaxLen(10).
			Comment("消费货币，如 USD、USDT"),

		field.Int8("consume_type").
			Default(1).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("消费类型 1:游戏 2:其他"),

		field.Time("consume_at").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("消费时间"),

		// 余额数据
		field.Int64("balance_before").
			Default(0).
			Comment("消费前余额（分）"),

		field.Int64("balance_after").
			Default(0).
			Comment("消费后余额（分）"),

		field.String("balance_currency").
			Optional().
			MaxLen(10).
			Comment("余额货币"),

		// 关联信息
		field.Int64("agent_id").
			Optional().
			Comment("所属代理ID"),

		field.Int64("wallet_id").
			Optional().
			Comment("钱包ID"),

		field.Int64("game_id").
			Optional().
			Comment("游戏ID"),

		field.Text("notes").
			Optional().
			Comment("备注"),
	}
}

// Indexes 定义模拟营业额表索引
func (Turnover) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("member_id", "consume_at").
			StorageKey("idx_turnover_user_consume_at"),
		index.Fields("agent_id", "consume_at").
			StorageKey("idx_turnover_agent_consume_at"),
	}
}

// Mixin 定义模拟营业额表公共字段
func (Turnover) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		mixins.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义模拟营业额表查询隔离
func (Turnover) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义模拟营业额表数据库注解
func (Turnover) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("模拟营业额表"),
		entsql.Annotation{
			Table: "turnover",
		},
	}
}
