package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberProfile 定义会员个人资料表结构
type MemberProfile struct {
	ent.Schema
}

// Fields 定义会员个人资料表字段
func (MemberProfile) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 会员归属 ====================

		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		// ==================== 会员形象 ====================

		field.String("avatar_url").
			NotEmpty().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("会员头像URL, 使用系统预设资源"),

		field.String("avatar_frame_url").
			NotEmpty().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Comment("会员头像框URL, 使用系统预设资源"),

		// ==================== 个人资料 ====================

		field.String("nickname").
			NotEmpty().
			MaxLen(100).
			Comment("会员昵称"),

		field.String("real_name_ciphertext").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Sensitive().
			Comment("会员真实姓名加密密文"),

		field.String("identity_number_ciphertext").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Sensitive().
			Comment("会员身份证件号码或CPF加密密文"),

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

		// ==================== 会员关联 ====================

		// 每个会员只允许有一条个人资料
		index.Fields("member_id").
			Unique().
			StorageKey("uk_member_profile_member"),

		// 按厅查询会员个人资料
		index.Fields("operator_code").
			StorageKey("idx_member_profile_operator"),
	}
}

// Mixin 定义会员个人资料表公共字段
func (MemberProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors 定义会员个人资料表查询隔离
func (MemberProfile) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks 定义会员个人资料表写入隔离
func (MemberProfile) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
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
