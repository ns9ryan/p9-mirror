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

// AgentApply 定义代理申请表结构。
type AgentApply struct {
	ent.Schema
}

// Fields 定义代理申请表字段
func (AgentApply) Fields() []ent.Field {
	return []ent.Field{
		field.String("account").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("申请代理账号"),

		field.String("realname").
			NotEmpty().
			MaxLen(100).
			Match(trimmedNonEmpty).
			Comment("真实姓名"),

		field.String("mobile").
			NotEmpty().
			MaxLen(32).
			Match(trimmedNonEmpty).
			Comment("手机号"),

		field.Int64("register_source").
			Range(1, 4).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("注册来源：1 PC 2 app 3 管理员 4 代理"),

		field.Int64("apply_share_rate").
			Range(0, 100).
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("申请占成比例，百分比，0~100"),

		field.String("member_account").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("使用会员账号"),

		field.Int64("member_id").
			Optional().
			Nillable().
			Comment("使用会员ID"),

		field.Int64("audit_status").
			Default(1).
			Range(1, 4).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("审核状态：1待审核 2已通过 3人工拒绝 4风控拒绝"),

		field.String("reject_reason").
			Optional().
			Nillable().
			MaxLen(500).
			Match(trimmedNonEmpty).
			Comment("拒绝理由"),

		field.String("apply_remark").
			Optional().
			Nillable().
			MaxLen(500).
			Match(trimmedNonEmpty).
			Comment("申请备注"),

		field.Time("register_at").
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("申请/注册时间"),

		field.Time("audited_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("审核时间"),

		field.String("auditor").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("审核人"),

		field.Int64("agent_id").
			Optional().
			Nillable().
			Comment("通过后创建的代理ID"),
	}
}

// Indexes 定义代理申请表索引
func (AgentApply) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "audit_status").
			StorageKey("idx_agent_apply_operator_status"),
		index.Fields("operator_code", "account").
			StorageKey("idx_agent_apply_operator_account"),
	}
}

// Mixin 定义代理申请表公共字段
func (AgentApply) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeMixin{},
	}
}

// Annotations 定义代理申请表数据库注解
func (AgentApply) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理申请表"),
		entsql.Annotation{
			Table: "agent_apply",
			Checks: map[string]string{
				"ck_agent_apply_operator_code":    "BTRIM(operator_code) <> '' AND operator_code = BTRIM(operator_code)",
				"ck_agent_apply_account":          "BTRIM(account) <> '' AND account = BTRIM(account)",
				"ck_agent_apply_realname":         "BTRIM(realname) <> '' AND realname = BTRIM(realname)",
				"ck_agent_apply_mobile":           "BTRIM(mobile) <> '' AND mobile = BTRIM(mobile)",
				"ck_agent_apply_register_source":  "register_source BETWEEN 1 AND 4",
				"ck_agent_apply_apply_share_rate": "apply_share_rate BETWEEN 0 AND 100",
				"ck_agent_apply_audit_status":     "audit_status BETWEEN 1 AND 4",
			},
		},
	}
}
