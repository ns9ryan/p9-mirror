package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AgentClosure 定义代理祖先与子孙的闭包关系，包含自身。
// depth=0 表示自己，depth=1 表示直属上下级。
type AgentClosure struct {
	ent.Schema
}

// Fields 定义代理闭包表字段
func (AgentClosure) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ancestor_id").
			Immutable().
			Comment("祖先agent ID"),

		field.Int64("descendant_id").
			Immutable().
			Comment("子孙agent ID"),

		field.Int64("depth").
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("0=自己，1=直属上级，越大越接近分站"),
	}
}

// Edges 定义代理闭包表关联关系
func (AgentClosure) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("ancestor", AgentUser.Type).
			Field("ancestor_id").
			Required().
			Unique().
			Immutable().
			StorageKey(edge.Symbol("fk_agent_closure_ancestor")).
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("descendant", AgentUser.Type).
			Field("descendant_id").
			Required().
			Unique().
			Immutable().
			StorageKey(edge.Symbol("fk_agent_closure_descendant")).
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

// Indexes 定义代理闭包表索引
func (AgentClosure) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("descendant_id", "depth").
			StorageKey("idx_agent_closure_descendant"),
		index.Fields("ancestor_id", "depth").
			StorageKey("idx_agent_closure_ancestor"),
	}
}

// Annotations 定义代理闭包表数据库注解。
// 复合主键是 (ancestor_id, descendant_id)。agent id 全局唯一，已能唯一确定一条闭包。
func (AgentClosure) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("ancestor_id", "descendant_id"),
		entsql.WithComments(true),
		schema.Comment("代理祖先-子孙闭包，含自己"),
		entsql.Annotation{
			Table: "agent_closure",
			Checks: map[string]string{
				"ck_agent_closure_depth": "(ancestor_id = descendant_id AND depth = 0) OR (ancestor_id <> descendant_id AND depth > 0)",
			},
		},
	}
}
