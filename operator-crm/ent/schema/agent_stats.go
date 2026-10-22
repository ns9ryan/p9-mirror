package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
)

// AgentStats 定义代理统计表结构。
// 直属只算下一级，所有下级不含自身。佣金金额按结算货币最小单位保存。
type AgentStats struct {
	ent.Schema
}

// Fields 定义代理统计表字段
func (AgentStats) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("agent_id").
			Immutable().
			Comment("被统计的agent ID"),

		field.Int64("direct_member_count").
			Default(0).
			NonNegative().
			Comment("直属会员数"),

		field.Int64("direct_agent_count").
			Default(0).
			NonNegative().
			Comment("直属代理数"),

		field.Int64("total_member_count").
			Default(0).
			NonNegative().
			Comment("所有会员数"),

		field.Int64("total_agent_count").
			Default(0).
			NonNegative().
			Comment("所有代理数"),

		field.Int64("total_commission_amount").
			Default(0).
			NonNegative().
			Comment("累计佣金（最小货币单位）"),

		field.Int64("unpaid_commission_amount").
			Default(0).
			NonNegative().
			Comment("未发放佣金（最小货币单位）"),

		field.Int64("commission_claim_count").
			Default(0).
			NonNegative().
			Comment("佣金领取次数"),

		field.Int64("visit_uv").
			Default(0).
			NonNegative().
			Comment("访问量UV"),
	}
}

// Edges 定义代理统计表关联关系
func (AgentStats) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent", AgentUser.Type).
			Ref("stats").
			Field("agent_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义代理统计表索引
func (AgentStats) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id").
			Unique().
			StorageKey("uk_agent_stats_agent"),
	}
}

// Mixin 定义代理统计表公共字段
func (AgentStats) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
	}
}

// Annotations 定义代理统计表数据库注解
func (AgentStats) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("agent统计表"),
		entsql.Annotation{
			Table: "agent_stats",
		},
	}
}
