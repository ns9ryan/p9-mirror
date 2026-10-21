package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberVipLevel 定义会员VIP等级表结构
type MemberVipLevel struct {
	ent.Schema
}

// Fields 定义会员VIP等级表字段
func (MemberVipLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("level").
			NonNegative().
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("VIP等级编号, 从0开始连续递增"),

		field.Int64("upgrade_deposit").
			NonNegative().
			Comment("晋级所需存款金额"),

		field.Int64("upgrade_water").
			NonNegative().
			Comment("晋级所需洗码量"),

		field.Int64("retention_deposit").
			NonNegative().
			Comment("保级所需存款金额"),

		field.Int64("retention_water").
			NonNegative().
			Comment("保级所需洗码量"),

		field.Int64("upgrade_reward").
			NonNegative().
			Default(0).
			Comment("晋级奖励金额"),
	}
}

// Indexes 定义会员VIP等级表索引
func (MemberVipLevel) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个厅的VIP等级编号不能重复
		index.Fields("operator_code", "level").
			Unique().
			StorageKey("uk_member_vip_level_operator_level"),
	}
}

// Mixin 定义会员VIP等级表公共字段
func (MemberVipLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors VIP等级查询隔离
func (MemberVipLevel) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks VIP等级写入隔离
func (MemberVipLevel) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
	}
}

// Annotations 定义会员VIP等级表数据库注解
func (MemberVipLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                    // 启用数据库字段注释
		schema.Comment("会员VIP等级表"),                   // 设置数据库表注释
		entsql.Annotation{Table: "member_vip_level"}, // 设置数据库表名
	}
}
