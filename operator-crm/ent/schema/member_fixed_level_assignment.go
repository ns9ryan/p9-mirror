package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberFixedLevelAssignment 定义会员固定层级分配表结构
type MemberFixedLevelAssignment struct {
	ent.Schema
}

// Fields 定义会员固定层级分配表字段
func (MemberFixedLevelAssignment) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		field.Int64("fixed_level_id").
			Positive().
			Immutable().
			Comment("固定层级ID"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("分配时间"),
	}
}

// Edges 定义会员固定层级分配表关联关系
func (MemberFixedLevelAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("fixed_level_assignments").
			Field("member_id").
			Unique().
			Required().
			Immutable(),

		edge.From("fixed_level", MemberFixedLevel.Type).
			Ref("member_assignments").
			Field("fixed_level_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员固定层级分配表索引
func (MemberFixedLevelAssignment) Indexes() []ent.Index {
	return []ent.Index{
		// 同一会员不能重复分配同一个固定层级
		index.Fields("member_id", "fixed_level_id").
			Unique().
			StorageKey("uk_member_fixed_level_assignment_member_level"),

		// 查询指定固定层级下的会员
		index.Fields("fixed_level_id").
			StorageKey("idx_member_fixed_level_assignment_level"),
	}
}

// Mixin 定义会员固定层级分配表公共字段
func (MemberFixedLevelAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
	}
}

// Annotations 定义会员固定层级分配表数据库注解
func (MemberFixedLevelAssignment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("会员固定层级分配表"),
		entsql.Annotation{Table: "member_fixed_level_assignment"},
	}
}
