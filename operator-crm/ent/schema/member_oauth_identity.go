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

// MemberOAuthIdentity 定义会员第三方授权身份表结构
type MemberOAuthIdentity struct {
	ent.Schema
}

// Fields 定义会员第三方授权身份表字段
func (MemberOAuthIdentity) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		field.String("provider").
			NotEmpty().
			MaxLen(32).
			Immutable().
			Comment("第三方授权平台编码"),

		field.String("provider_user_id").
			NotEmpty().
			MaxLen(255).
			Immutable().
			Comment("第三方平台用户唯一标识"),

		field.String("provider_account").
			NotEmpty().
			MaxLen(255).
			Optional().
			Nillable().
			Comment("第三方平台展示账号"),
	}
}

// Edges 定义会员第三方授权身份表关联关系
func (MemberOAuthIdentity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("oauth_identities").
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员第三方授权身份表索引
func (MemberOAuthIdentity) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个会员在同一平台只能绑定一个账号
		index.Fields("member_id", "provider").
			Unique().
			StorageKey("uk_member_oauth_identity_member_provider"),

		// 同一个厅的同一第三方账号只能绑定一个会员
		index.Fields("operator_code", "provider", "provider_user_id").
			Unique().
			StorageKey("uk_member_oauth_identity_operator_provider_user"),
	}
}

// Mixin 定义会员第三方授权身份表公共字段
func (MemberOAuthIdentity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员第三方授权身份表数据库注解
func (MemberOAuthIdentity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                         // 启用数据库字段注释
		schema.Comment("会员第三方授权身份表"),                      // 设置数据库表注释
		entsql.Annotation{Table: "member_oauth_identity"}, // 设置数据库表名
	}
}
