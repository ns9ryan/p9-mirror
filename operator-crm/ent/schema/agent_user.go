package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
)

// AgentUser 定义代理主表结构。
// 顶层代理 parent_id、top_id 为空，level=0。
type AgentUser struct {
	ent.Schema
}

// Fields 定义代理主表字段
func (AgentUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("account").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("登录账号"),

		field.String("code").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("代理编码"),

		field.Int64("parent_id").
			Optional().
			Nillable().
			Comment("直接上级代理ID，顶层为空"),

		field.Int64("top_id").
			Optional().
			Nillable().
			Comment("顶层代理ID，顶层为空"),

		field.Int64("level").
			Default(0).
			Range(0, 7).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("树层级，顶层为0，最大7"),

		field.Int64("agent_type").
			Range(1, 2).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("1现金代理 2信用代理"),

		field.String("agent_line_code").
			NotEmpty().
			MaxLen(32).
			Match(trimmedNonEmpty).
			Comment("代理子线路编码"),

		field.Int64("share_rate").
			Default(0).
			Range(0, 100).
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("当前占成比例，百分比，0~100（100=100%）"),

		field.Int64("status").
			Default(1).
			Range(1, 3).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("1正常 2冻结 3停用"),

		field.Int64("payout_restricted").
			Default(1).
			Range(1, 3).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("1不限制出款 2限制本人 3限制全线"),

		field.Int64("register_source").
			Range(1, 4).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("1 PC注册 2 app注册 3管理员添加 4代理添加"),

		field.String("promo_code").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("对外推广码"),

		field.String("login_password_hash").
			Sensitive().
			NotEmpty().
			MaxLen(255).
			Match(trimmedNonEmpty).
			Comment("登录密码哈希"),

		field.String("withdraw_password_hash").
			Sensitive().
			Optional().
			Nillable().
			MaxLen(255).
			Match(trimmedNonEmpty).
			Comment("提现密码哈希"),

		field.String("id_document_no").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("CPF/身份证件号"),

		field.String("salt").
			Sensitive().
			NotEmpty().
			MaxLen(16).
			Comment("令牌盐"),

		field.String("remark").
			Optional().
			Nillable().
			MaxLen(1000).
			Match(trimmedNonEmpty).
			Comment("备注"),

		field.Time("register_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("账号开通时间"),

		field.String("register_ip").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Comment("注册IP"),

		field.Time("last_login_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("最后登录时间"),

		field.String("last_login_ip").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Comment("最后登录IP"),

		field.Int64("login_count").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("登录次数"),

		field.String("last_login_udid").
			Optional().
			Nillable().
			MaxLen(128).
			Comment("最后登录设备码"),
	}
}

// Edges 定义代理主表关联关系
func (AgentUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", AgentUser.Type).
			StorageKey(edge.Symbol("fk_agent_user_parent")).
			Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
		edge.From("parent", AgentUser.Type).
			Ref("children").
			Field("parent_id").
			Unique(),

		edge.To("line_agents", AgentUser.Type).
			StorageKey(edge.Symbol("fk_agent_user_top")).
			Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
		edge.From("top", AgentUser.Type).
			Ref("line_agents").
			Field("top_id").
			Unique(),

		edge.To("profile", AgentProfile.Type).
			Unique().
			StorageKey(edge.Symbol("fk_agent_profile_agent")).
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("stats", AgentStats.Type).
			Unique().
			StorageKey(edge.Symbol("fk_agent_stats_agent")).
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),

		edge.To("descendants", AgentUser.Type).
			Through("closures", AgentClosure.Type).
			StorageKey(
				edge.Columns("ancestor_id", "descendant_id"),
				edge.Symbols("fk_agent_closure_ancestor", "fk_agent_closure_descendant"),
			),
		edge.From("ancestors", AgentUser.Type).
			Ref("descendants").
			Through("ancestor_closures", AgentClosure.Type),
	}
}

// Indexes 定义代理主表索引
func (AgentUser) Indexes() []ent.Index {
	return []ent.Index{
		// 闭包表复合外键依赖此唯一约束
		index.Fields("id", "operator_code").
			Unique().
			StorageKey("uk_agent_user_id_operator"),
		index.Fields("operator_code", "account").
			Unique().
			StorageKey("uk_agent_user_operator_account"),
		index.Fields("operator_code", "code").
			Unique().
			StorageKey("uk_agent_user_operator_code"),
		index.Fields("operator_code", "promo_code").
			Unique().
			StorageKey("uk_agent_user_operator_promo"),
		index.Fields("operator_code", "parent_id").
			StorageKey("idx_agent_user_operator_parent"),
	}
}

// Mixin 定义代理主表公共字段
func (AgentUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeMixin{},
	}
}

// Annotations 定义代理主表数据库注解
func (AgentUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理主表"),
		entsql.Annotation{
			Table: "agent_user",
			Checks: map[string]string{
				"ck_agent_user_operator_code":          "BTRIM(operator_code) <> '' AND operator_code = BTRIM(operator_code)",
				"ck_agent_user_account":                "BTRIM(account) <> '' AND account = BTRIM(account)",
				"ck_agent_user_code":                   "BTRIM(code) <> '' AND code = BTRIM(code)",
				"ck_agent_user_agent_line_code":        "BTRIM(agent_line_code) <> '' AND agent_line_code = BTRIM(agent_line_code)",
				"ck_agent_user_promo_code":             "BTRIM(promo_code) <> '' AND promo_code = BTRIM(promo_code)",
				"ck_agent_user_login_password_hash":    "BTRIM(login_password_hash) <> ''",
				"ck_agent_user_withdraw_password_hash": "withdraw_password_hash IS NULL OR BTRIM(withdraw_password_hash) <> ''",
				"ck_agent_user_id_document_no":         "id_document_no IS NULL OR BTRIM(id_document_no) <> ''",
				"ck_agent_user_remark":                 "remark IS NULL OR BTRIM(remark) <> ''",
				"ck_agent_user_top_self":               "top_id IS DISTINCT FROM id",
				"ck_agent_user_top_id":                 "(top_id IS NULL AND level = 0) OR (top_id IS NOT NULL AND level BETWEEN 1 AND 7)",
				"ck_agent_user_parent_self":            "parent_id IS DISTINCT FROM id",
				"ck_agent_user_tree":                   "(parent_id IS NULL AND level = 0) OR (parent_id IS NOT NULL AND level BETWEEN 1 AND 7)",
				"ck_agent_user_agent_type":             "agent_type IN (1, 2)",
				"ck_agent_user_share_rate":             "share_rate BETWEEN 0 AND 100",
				"ck_agent_user_status":                 "status IN (1, 2, 3)",
				"ck_agent_user_payout_restricted":      "payout_restricted IN (1, 2, 3)",
				"ck_agent_user_register_source":        "register_source IN (1, 2, 3, 4)",
				"ck_agent_user_login_count":            "login_count >= 0",
			},
		},
	}
}
