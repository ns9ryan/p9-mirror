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

// AgentConfig 定义运营商代理配置。
// 每个运营商一行。指定代理只保存 agent_user.id，不建外键。
type AgentConfig struct {
	ent.Schema
}

// Fields 定义代理配置表字段
func (AgentConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("max_share_rate").
			Range(0, 100).
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("最高占成比例，百分比，0~100（100=100%）"),

		field.Int64("apply_min_share_rate").
			Range(0, 100).
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("申请最低占成比例，百分比，0~100（100=100%）"),

		field.Int64("apply_max_share_rate").
			Range(0, 100).
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("申请最高占成比例，百分比，0~100（100=100%）"),

		field.Int64("cash_max_member_count").
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("代理最大会员数（现金）"),

		field.Int64("cash_max_agent_per_ip").
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("IP最大可注册代理数（现金）"),

		field.Int64("cash_production_agent_id").
			Optional().
			Nillable().
			Comment("现金正式线路指定代理 CASH_PRODUCTION"),

		field.Int64("cash_demo_agent_id").
			Optional().
			Nillable().
			Comment("现金试玩线路指定代理 CASH_DEMO"),

		field.Int64("cash_test_agent_id").
			Optional().
			Nillable().
			Comment("现金测试线路指定代理 CASH_TEST"),

		field.Int64("credit_demo_agent_id").
			Optional().
			Nillable().
			Comment("信誉试玩线路指定代理 CREDIT_DEMO"),

		field.Int64("credit_test_agent_id").
			Optional().
			Nillable().
			Comment("信誉测试线路指定代理 CREDIT_TEST"),
	}
}

// Indexes 定义代理配置表索引
func (AgentConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code").
			Unique().
			StorageKey("uk_agent_config_operator"),
	}
}

// Mixin 定义代理配置表公共字段
func (AgentConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义代理配置表查询隔离
func (AgentConfig) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义代理配置表数据库注解
func (AgentConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理配置表"),
		entsql.Annotation{
			Table: "agent_config",
		},
	}
}
