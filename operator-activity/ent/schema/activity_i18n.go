package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"
)

// ActivityI18n 定义活动配置国际化表结构。
type ActivityI18n struct {
	ent.Schema
}

// Fields 定义活动国际化表字段
func (ActivityI18n) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").
			Immutable().
			NonNegative().
			Comment("活动ID"),

		field.String("lang").
			MaxLen(16).
			Comment("语言码"),

		field.String("name").
			MaxLen(128).
			Default("").
			Comment("活动名称"),

		field.Text("rule").
			MaxLen(4096).
			Default("").
			Comment("活动规则"),

		field.Text("description").
			MaxLen(1024).
			Default("").
			Comment("活动说明"),

		field.String("image").
			MaxLen(255).
			Default("").
			Comment("活动图片 URL"),
	}
}

// Edges 定义活动国际化表关联关系
func (ActivityI18n) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", Activity.Type).
			Ref("i18ns").
			Field("activity_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义活动国际化表索引
func (ActivityI18n) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "lang", "activity_id").Unique().
			StorageKey("idx_activity_i18n_operator_lang_activity"),
	}
}

// Mixin 定义活动国际化表公共字段
func (ActivityI18n) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动国际化表查询隔离
func (ActivityI18n) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动国际化表数据库注解
func (ActivityI18n) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动配置国际化表"),
		entsql.Annotation{
			Table: "activity_i18n",
		},
	}
}
