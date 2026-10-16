package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberFixedLevel 定义会员固定层级表结构
type MemberFixedLevel struct {
	ent.Schema
}

// Fields 定义会员固定层级表字段
func (MemberFixedLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("固定层级名称"),

		field.String("description").
			Optional().
			MaxLen(255).
			Comment("固定层级说明"),

		field.Int64("withdrawal_limit").
			Optional().
			Nillable().
			Positive().
			Comment("单笔提款金额上限, 为空表示不额外限制"),
	}
}

// Edges 定义会员固定层级表关联关系
func (MemberFixedLevel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("member_assignments", MemberFixedLevelAssignment.Type),
	}
}

// Indexes 定义会员固定层级表索引
func (MemberFixedLevel) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个厅的固定层级名称不能重复
		index.Fields("operator_code", "name").
			Unique().
			StorageKey("uk_member_fixed_level_operator_name"),
	}
}

// Mixin 定义会员固定层级表公共字段
func (MemberFixedLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员固定层级表数据库注解
func (MemberFixedLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                      // 启用数据库字段注释
		schema.Comment("会员固定层级表"),                      // 设置数据库表注释
		entsql.Annotation{Table: "member_fixed_level"}, // 设置数据库表名
	}
}
