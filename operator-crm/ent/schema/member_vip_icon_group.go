package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberVipIconGroup 定义会员VIP图标组表结构
type MemberVipIconGroup struct {
	ent.Schema
}

// Fields 定义会员VIP图标组表字段
func (MemberVipIconGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("group_no").
			NonNegative().
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("图标组编号, 每10个VIP等级共用一组"),

		field.String("default_icon_url").
			NotEmpty().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("系统默认图标URL"),

		field.String("icon_url").
			NotEmpty().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("当前使用的图标URL"),
	}
}

// Indexes 定义会员VIP图标组表索引
func (MemberVipIconGroup) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个厅的图标组编号不能重复
		index.Fields("operator_code", "group_no").
			Unique().
			StorageKey("uk_member_vip_icon_group_operator_group"),
	}
}

// Mixin 定义会员VIP图标组表公共字段
func (MemberVipIconGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员VIP图标组表数据库注解
func (MemberVipIconGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                         // 启用数据库字段注释
		schema.Comment("会员VIP图标组表"),                       // 设置数据库表注释
		entsql.Annotation{Table: "member_vip_icon_group"}, // 设置数据库表名
	}
}
