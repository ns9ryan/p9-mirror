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

// AgentLevel 定义代理等级表结构。
// 同一运营商每个等级一条，level 与代理主表层级范围一致。
type AgentLevel struct {
	ent.Schema
}

// Fields 定义代理等级表字段
func (AgentLevel) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("level").
			Range(1, 7).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("等级，1~7"),

		field.String("name").
			NotEmpty().
			MaxLen(32).
			Match(trimmedNonEmpty).
			Comment("等级名称"),
		field.String("alias").
			NotEmpty().
			MaxLen(32).
			Match(trimmedNonEmpty).
			Comment("等级别名"),
	}
}

// Indexes 定义代理等级表索引
func (AgentLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "level").
			Unique().
			StorageKey("uk_agent_level_operator_level"),
	}
}

// Mixin 定义代理等级表公共字段
func (AgentLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeMixin{},
	}
}

// Interceptors 定义代理等级表查询隔离
func (AgentLevel) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义代理等级表数据库注解
func (AgentLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("代理等级表"),
		entsql.Annotation{
			Table: "agent_level",
		},
	}
}
