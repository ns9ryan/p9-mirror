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
)

// AgentDepositOrderAudit 定义代理存款订单审核记录表结构。
type AgentDepositOrderAudit struct {
	ent.Schema
}

// Fields 定义代理存款订单审核记录表字段
func (AgentDepositOrderAudit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("order_id").
			Comment("关联的存款订单ID"),

		field.Int8("action").
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("操作类型 1=approve 2=reject 3=freeze 4=unfreeze"),

		field.Int8("stage").
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("审核阶段 1=一审 2=二审 0=非审核操作(冻结/解冻)"),

		field.Int8("audit_utype").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("审核用户类型"),

		field.Int64("audit_uid").
			Optional().
			Comment("审核用户ID"),

		field.Time("audit_at").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("审核时间"),

		field.String("audit_ip").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Comment("审核IP"),

		field.Text("audit_notes").
			Optional().
			Comment("审核备注"),

		field.String("audit_username").
			Optional().
			MaxLen(64).
			Comment("审核人用户名，approve/reject 时从 claims 写入"),
	}
}

// Indexes 定义代理存款订单审核记录表索引
func (AgentDepositOrderAudit) Indexes() []ent.Index {
	return []ent.Index{
		// 同一订单可多次冻结/解冻，不设唯一约束，按 audit_at 排序还原操作时序
		index.Fields("order_id").
			StorageKey("idx_agent_deposit_order_audit_order_id"),
	}
}

// Mixin 定义代理存款订单审核记录表公共字段
func (AgentDepositOrderAudit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义代理存款订单审核记录表查询隔离
func (AgentDepositOrderAudit) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义代理存款订单审核记录表数据库注解
func (AgentDepositOrderAudit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理存款订单审核记录表"),
		entsql.Annotation{
			Table: "agent_deposit_order_audit",
		},
	}
}
