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

// AgentOpLog 定义代理资料变更操作日志表结构。
type AgentOpLog struct {
	ent.Schema
}

// Fields 定义代理资料变更操作日志表字段
func (AgentOpLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("agent_id").
			Comment("被操作代理ID"),

		field.Int64("operator_id").
			Optional().
			Nillable().
			Immutable().
			Comment("操作人用户ID"),

		field.String("operator_account").
			Default("").
			MaxLen(64).
			Immutable().
			Comment("操作用户名"),

		field.String("action").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("操作项目 code"),

		field.String("before_value").
			Optional().
			Nillable().
			MaxLen(2000).
			Immutable().
			Comment("变更前"),

		field.String("after_value").
			Optional().
			Nillable().
			MaxLen(2000).
			Immutable().
			Comment("变更后"),

		field.String("client_ip").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Comment("客户端IP"),

		field.String("login_device").
			Optional().
			Nillable().
			MaxLen(512).
			Match(trimmedNonEmpty).
			Comment("登录设备，User-Agent"),

		field.String("udid").
			Optional().
			Nillable().
			MaxLen(128).
			Match(trimmedNonEmpty).
			Comment("设备码"),

		field.Time("operated_at").
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("操作时间"),

		field.Int64("result").
			Range(1, 2).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("结果：1成功 2失败"),

		field.String("fail_reason").
			Optional().
			Nillable().
			MaxLen(512).
			Comment("失败原因，成功为空"),
	}
}

// Indexes 定义代理资料变更操作日志表索引
func (AgentOpLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "agent_id", "operated_at").
			StorageKey("idx_agent_op_log_operator_agent_operated_at"),
		index.Fields("operator_code", "operated_at").
			StorageKey("idx_agent_op_log_operator_operated_at"),
	}
}

// Mixin 定义代理资料变更操作日志表公共字段
func (AgentOpLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Annotations 定义代理资料变更操作日志表数据库注解
func (AgentOpLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理资料变更操作日志表"),
		entsql.Annotation{
			Table: "agent_op_log",
		},
	}
}
