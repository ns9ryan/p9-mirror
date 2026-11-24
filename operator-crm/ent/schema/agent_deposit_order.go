package schema

import (
	"context"
	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
)

// AgentDepositOrder 定义代理存款订单表结构。
type AgentDepositOrder struct {
	ent.Schema
}

// Fields 定义代理存款订单表字段
func (AgentDepositOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int8("agent_type").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("代理模式 1.现金 2.信用代理"),

		//代理子线路编码
		field.String("agent_line_code").
			NotEmpty().
			MaxLen(32).
			Immutable().
			Comment("代理子线路编码"),

		field.Int64("agent_id").
			Optional().
			Comment("代理ID"),

		field.Int8("status").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("订单状态 1:待审核 2:冻结 3:解冻 4:通过 5:拒绝 11:待一审核 12:待二审 "),

		field.Int64("deposit_amount").
			Optional().
			Comment("存款金额"),

		field.String("deposit_currency").
			Optional().
			MaxLen(10).
			Comment("存款货币"),

		field.Int8("audit_stages").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("需要审核的阶段数"),

		field.Float("fee_rate").
			Optional().
			Default(0.0).
			Comment("费率"),

		field.Float("bonus_rate").
			Optional().
			Comment("奖金率"),

		field.Int64("fee_amount").
			Default(0).
			Comment("手续费金额"),

		field.Int64("bonus_amount").
			Default(0).
			Comment("奖金金额"),

		field.Int64("received_amount").
			Default(0).
			Comment("到账金额 = 存款金额 + 奖金金额"),

		field.Int64("actual_received_amount").
			Default(0).
			Comment("实际到账金额 = 到账金额 - 手续费金额"),

		field.Int64("member_id").
			Optional().
			Comment("会员ID"),

		field.Int8("fee_mode").
			Default(1).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("费率模式"),

		field.Float("rollover_multiplier").
			Default(0).
			Comment("洗码倍数"),

		field.Int8("deposit_type").
			Default(3).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("存款方式  1:加密货币 2:人工入账"),

		field.Text("notes").
			Optional().
			Comment("备注"),

		field.String("order_no").
			Optional().
			MaxLen(64).
			Comment("订单号"),

		field.String("site_currency").
			Optional().
			MaxLen(10).
			Comment("站点货币，如 USD、USDT"),

		field.String("third_party_order_no").
			Optional().
			MaxLen(128).
			Comment("三方订单号"),

		field.Text("payment_receive_info").
			Optional().
			Comment("存款信息（收款方）"),

		field.Text("payment_send_info").
			Optional().
			Comment("存款信息（付款方）"),

		// ==================== 操作人 ====================

		field.Int64("creator_id").
			Default(0).
			Comment("创建人用户ID"),

		field.Int8("creator_type").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("创建人类型 1:总网 2:分站"),

		field.Int64("updater_id").
			Default(0).
			Comment("最后更新人用户ID"),

		field.Int8("updater_type").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("最后更新人类型 1:总网 2:分站"),

		// ==================== 审核领取 ====================

		field.Int64("reviewer_id").
			Default(0).
			Comment("当前领取审核人用户ID，0=未被领取"),

		field.Int8("reviewer_utype").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("领取人类型 1:总网 2:分站"),

		field.Time("claimed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("领取审核时间"),
	}
}

// Indexes 定义代理存款订单表索引
func (AgentDepositOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id", "status").
			StorageKey("idx_agent_deposit_order_agent_status"),
		index.Fields("created_at").
			StorageKey("idx_agent_deposit_order_created_at"),
		index.Fields("member_id").
			StorageKey("idx_agent_deposit_order_member"),
		index.Fields("reviewer_id", "status").
			StorageKey("idx_agent_deposit_order_reviewer_status"),
	}
}

// Mixin 定义代理存款订单表公共字段
func (AgentDepositOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		mixins.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义代理存款订单表查询隔离
func (AgentDepositOrder) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义代理存款订单表数据库注解
func (AgentDepositOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("存款订单表"),
		entsql.Annotation{
			Table: "agent_deposit_order",
		},
	}
}
