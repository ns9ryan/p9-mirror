package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberAutoLevel 定义会员自动层级表结构
type MemberAutoLevel struct {
	ent.Schema
}

// Fields 定义会员自动层级表字段
func (MemberAutoLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("自动层级名称"),

		field.Int64("min_valid_deposit").
			NonNegative().
			Comment("晋级要求的累计有效存款金额"),

		field.Int64("min_valid_turnover").
			NonNegative().
			Comment("晋级要求的累计有效洗码量"),
	}
}

// Indexes 定义会员自动层级表索引
func (MemberAutoLevel) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个厅不允许存在两个晋升条件完全相同的层级
		index.Fields("operator_code", "min_valid_deposit", "min_valid_turnover").
			Unique().
			StorageKey("uk_member_auto_level_threshold"),
	}
}

// Mixin 定义会员自动层级表公共字段
func (MemberAutoLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员自动层级表数据库注解
func (MemberAutoLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                     // 启用数据库字段注释
		schema.Comment("会员自动层级表"),                     // 设置数据库表注释
		entsql.Annotation{Table: "member_auto_level"}, // 设置数据库表名
	}
}
