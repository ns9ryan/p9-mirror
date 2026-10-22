package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
)

// AgentProfile 定义代理资料表结构。
// 登录账号、推广码和占成仍保存在代理主表。
type AgentProfile struct {
	ent.Schema
}

// Fields 定义代理资料表字段
func (AgentProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("agent_id").
			Immutable().
			Comment("所属agent ID"),

		field.Int64("gender").
			Default(0).
			Range(0, 2).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("性别：0保密，1男，2女"),

		field.Time("birth_date").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "date",
			}).
			Comment("生日"),

		field.String("avatar_url").
			Optional().
			Nillable().
			MaxLen(255).
			Match(trimmedNonEmpty).
			Comment("头像访问地址"),

		field.String("nickname").
			Optional().
			Nillable().
			MaxLen(100).
			Match(trimmedNonEmpty).
			Comment("昵称"),

		field.String("realname").
			Optional().
			Nillable().
			MaxLen(100).
			Match(trimmedNonEmpty).
			Comment("真实姓名"),

		field.String("mobile").
			Optional().
			Nillable().
			MaxLen(32).
			Match(trimmedNonEmpty).
			Comment("手机号码"),

		field.String("email").
			Optional().
			Nillable().
			MaxLen(255).
			Match(trimmedNonEmpty).
			Comment("邮箱"),

		field.String("id_document_no").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("CPF/身份证件号"),

		field.String("qq").
			Optional().
			Nillable().
			MaxLen(32).
			Match(trimmedNonEmpty).
			Comment("QQ"),

		field.String("line").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("LINE"),

		field.String("wechat").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("微信"),

		field.String("whatsapp").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("WhatsApp"),

		field.String("telegram").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("Telegram"),

		field.String("facebook").
			Optional().
			Nillable().
			MaxLen(128).
			Match(trimmedNonEmpty).
			Comment("Facebook"),

		field.String("twitter").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("Twitter"),

		field.String("zalo").
			Optional().
			Nillable().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Comment("Zalo"),
	}
}

// Edges 定义代理资料表关联关系
func (AgentProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent", AgentUser.Type).
			Ref("profile").
			Field("agent_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义代理资料表索引
func (AgentProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id").
			Unique().
			StorageKey("uk_agent_profile_agent"),
	}
}

// Mixin 定义代理资料表公共字段
func (AgentProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeMixin{},
	}
}

// Interceptors 定义代理资料表查询隔离
func (AgentProfile) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义代理资料表数据库注解
func (AgentProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("agent资料表"),
		entsql.Annotation{
			Table: "agent_profile",
		},
	}
}
