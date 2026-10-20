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

// AgentLoginLog 定义代理登录日志表结构。
type AgentLoginLog struct {
	ent.Schema
}

// Fields 定义代理登录日志表字段
func (AgentLoginLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("agent_id").
			Optional().
			Nillable().
			Comment("代理ID，账号不存在时为空"),

		field.Int64("parent_id").
			Optional().
			Nillable().
			Comment("父代理ID，账号不存在时为空"),

		field.String("account").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("登录账号"),

		field.Int64("login_type").
			Range(1, 1).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("登录类型：1账号密码"),

		field.String("login_ip").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Comment("登录IP"),

		field.String("login_location").
			Optional().
			Nillable().
			MaxLen(128).
			Match(trimmedNonEmpty).
			Comment("登录地点"),

		field.String("udid").
			Optional().
			Nillable().
			MaxLen(128).
			Match(trimmedNonEmpty).
			Comment("设备码"),

		field.String("login_device").
			Optional().
			Nillable().
			MaxLen(512).
			Match(trimmedNonEmpty).
			Comment("登录设备，User-Agent"),

		field.Int64("device_type").
			Range(1, 5).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("设备类型：1 PC网页 2 H5 3 Android 4 iOS 5 PC客户端"),

		field.Time("login_at").
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("登录时间"),

		field.Int64("result").
			Range(1, 4).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("登录结果：1成功 2密码错误 3账号停用 4账号不存在"),
	}
}

// Indexes 定义代理登录日志表索引
func (AgentLoginLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "login_at").
			StorageKey("idx_agent_login_log_operator_login_at"),
		index.Fields("agent_id").
			StorageKey("idx_agent_login_log_agent_id"),
		index.Fields("operator_code", "account").
			StorageKey("idx_agent_login_log_operator_account"),
		index.Fields("operator_code", "parent_id").
			StorageKey("idx_agent_login_log_operator_parent_id"),
		index.Fields("login_ip").
			StorageKey("idx_agent_login_log_login_ip"),
		index.Fields("udid").
			StorageKey("idx_agent_login_log_udid"),
	}
}

// Mixin 定义代理登录日志表公共字段
func (AgentLoginLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeMixin{},
	}
}

// Annotations 定义代理登录日志表数据库注解
func (AgentLoginLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理登录日志表"),
		entsql.Annotation{
			Table: "agent_login_log",
			Checks: map[string]string{
				"ck_agent_login_log_operator_code":  "BTRIM(operator_code) <> '' AND operator_code = BTRIM(operator_code)",
				"ck_agent_login_log_account":        "BTRIM(account) <> '' AND account = BTRIM(account)",
				"ck_agent_login_log_login_type":     "login_type = 1",
				"ck_agent_login_log_udid":           "udid IS NULL OR (BTRIM(udid) <> '' AND udid = BTRIM(udid))",
				"ck_agent_login_log_login_location": "login_location IS NULL OR (BTRIM(login_location) <> '' AND login_location = BTRIM(login_location))",
				"ck_agent_login_log_login_device":   "login_device IS NULL OR (BTRIM(login_device) <> '' AND login_device = BTRIM(login_device))",
			},
		},
	}
}
