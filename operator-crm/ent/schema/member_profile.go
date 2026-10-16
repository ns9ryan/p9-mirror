package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberProfile 定义会员个人资料表结构
type MemberProfile struct {
	ent.Schema
}

// Fields 定义会员个人资料表字段
func (MemberProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		// ==================== 会员形象 ====================

		field.String("nickname").
			NotEmpty().
			MaxLen(100).
			Comment("会员昵称"),

		field.String("avatar_url").
			NotEmpty().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("会员头像URL"),

		field.String("avatar_frame_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("会员头像框URL"),

		// ==================== 个人资料 ====================

		field.String("real_name").
			NotEmpty().
			MaxLen(100).
			Optional().
			Nillable().
			Comment("会员真实姓名"),

		field.String("identity_number").
			NotEmpty().
			MaxLen(100).
			Optional().
			Nillable().
			Sensitive().
			Comment("会员身份证件号码或CPF"),

		field.Time("birthday").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "date",
			}).
			Comment("会员生日"),

		field.Int64("gender").
			Default(0).
			Range(0, 3).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("性别: 0未设置, 1男, 2女, 3保密"),

		field.String("region_code").
			NotEmpty().
			MaxLen(2).
			Optional().
			Nillable().
			Comment("会员国家或地区编码"),

		// ==================== 后台管理 ====================

		field.String("remark").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("后台会员备注"),
	}
}

// Edges 定义会员个人资料表关联关系
func (MemberProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("profile").
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员个人资料表索引
func (MemberProfile) Indexes() []ent.Index {
	return []ent.Index{
		// 每个会员只允许有一条个人资料
		index.Fields("member_id").
			Unique().
			StorageKey("uk_member_profile_member"),
	}
}

// Mixin 定义会员个人资料表公共字段
func (MemberProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员个人资料表数据库注解
func (MemberProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                  // 启用数据库字段注释
		schema.Comment("会员个人资料表"),                  // 设置数据库表注释
		entsql.Annotation{Table: "member_profile"}, // 设置数据库表名
	}
}
