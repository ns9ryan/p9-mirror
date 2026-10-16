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
	}
}

// Annotations 定义代理资料表数据库注解
func (AgentProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("agent资料表"),
		entsql.Annotation{
			Table: "agent_profile",
			Checks: map[string]string{
				"ck_agent_profile_gender":     "gender IN (0, 1, 2)",
				"ck_agent_profile_birth_date": "birth_date IS NULL OR birth_date <= CURRENT_DATE",
				"ck_agent_profile_avatar_url": "avatar_url IS NULL OR (BTRIM(avatar_url) <> '' AND avatar_url = BTRIM(avatar_url))",
				"ck_agent_profile_nickname":   "nickname IS NULL OR (BTRIM(nickname) <> '' AND nickname = BTRIM(nickname))",
				"ck_agent_profile_realname":   "realname IS NULL OR (BTRIM(realname) <> '' AND realname = BTRIM(realname))",
				"ck_agent_profile_mobile":     "mobile IS NULL OR (BTRIM(mobile) <> '' AND mobile = BTRIM(mobile))",
				"ck_agent_profile_email":      "email IS NULL OR (BTRIM(email) <> '' AND email = BTRIM(email))",
				"ck_agent_profile_qq":         "qq IS NULL OR (BTRIM(qq) <> '' AND qq = BTRIM(qq))",
				"ck_agent_profile_line":       "line IS NULL OR (BTRIM(line) <> '' AND line = BTRIM(line))",
				"ck_agent_profile_wechat":     "wechat IS NULL OR (BTRIM(wechat) <> '' AND wechat = BTRIM(wechat))",
				"ck_agent_profile_whatsapp":   "whatsapp IS NULL OR (BTRIM(whatsapp) <> '' AND whatsapp = BTRIM(whatsapp))",
				"ck_agent_profile_telegram":   "telegram IS NULL OR (BTRIM(telegram) <> '' AND telegram = BTRIM(telegram))",
				"ck_agent_profile_facebook":   "facebook IS NULL OR (BTRIM(facebook) <> '' AND facebook = BTRIM(facebook))",
				"ck_agent_profile_twitter":    "twitter IS NULL OR (BTRIM(twitter) <> '' AND twitter = BTRIM(twitter))",
				"ck_agent_profile_zalo":       "zalo IS NULL OR (BTRIM(zalo) <> '' AND zalo = BTRIM(zalo))",
			},
		},
	}
}
