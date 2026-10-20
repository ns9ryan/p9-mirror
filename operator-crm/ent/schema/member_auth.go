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

// MemberAuth 定义会员认证信息表结构
type MemberAuth struct {
	ent.Schema
}

// Fields 定义会员认证信息表字段
func (MemberAuth) Fields() []ent.Field {
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

		// ==================== 登录身份 ====================

		field.String("username").
			NotEmpty().
			MaxLen(64).
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			Comment("会员登录账号, 保留原始大小写"),

		field.String("phone_ciphertext").
			NotEmpty().
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Sensitive().
			Comment("手机号加密密文"),

		field.String("phone_lookup_hash").
			NotEmpty().
			MinLen(64).
			MaxLen(64).
			Optional().
			Nillable().
			Sensitive().
			Comment("手机号检索摘要, 使用HMAC-SHA256生成"),

		field.String("email_ciphertext").
			NotEmpty().
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Sensitive().
			Comment("邮箱加密密文"),

		field.String("email_lookup_hash").
			NotEmpty().
			MinLen(64).
			MaxLen(64).
			Optional().
			Nillable().
			Sensitive().
			Comment("邮箱检索摘要, 使用HMAC-SHA256生成"),

		// ==================== 密码凭证 ====================

		field.String("login_password_hash").
			NotEmpty().
			MaxLen(255).
			Optional().
			Nillable().
			Sensitive().
			Comment("登录密码单向哈希"),

		field.String("withdrawal_password_hash").
			NotEmpty().
			MaxLen(255).
			Optional().
			Nillable().
			Sensitive().
			Comment("提款密码单向哈希"),

		// ==================== 验证失败次数 ====================

		field.Int64("sms_code_error_count").
			NonNegative().
			Default(0).
			Comment("短信验证码错误次数"),

		field.Int64("email_code_error_count").
			NonNegative().
			Default(0).
			Comment("邮箱验证码错误次数"),

		field.Int64("login_password_error_count").
			NonNegative().
			Default(0).
			Comment("登录密码错误次数"),

		field.Int64("withdrawal_password_error_count").
			NonNegative().
			Default(0).
			Comment("提款密码错误次数"),
	}
}

// Edges 定义会员认证信息表关联关系
func (MemberAuth) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("member", Member.Type).
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员认证信息表索引
func (MemberAuth) Indexes() []ent.Index {
	return []ent.Index{

		// ==================== 会员关联 ====================

		// 每个会员只能有一条认证信息
		index.Fields("member_id").
			Unique().
			StorageKey("uk_member_auth_member"),

		// ==================== 登录身份唯一约束 ====================

		// 同一个厅的登录账号唯一且不区分大小写
		index.Fields("operator_code", "username").
			Unique().
			StorageKey("uk_member_auth_operator_username"),

		// 同一个厅的手机号唯一
		index.Fields("operator_code", "phone_lookup_hash").
			Unique().
			StorageKey("uk_member_auth_operator_phone_hash"),

		// 同一个厅的邮箱唯一
		index.Fields("operator_code", "email_lookup_hash").
			Unique().
			StorageKey("uk_member_auth_operator_email_hash"),
	}
}

// Mixin 定义会员认证信息表公共字段
func (MemberAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors 定义会员认证信息表查询隔离
func (MemberAuth) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks 定义会员认证信息表写入隔离
func (MemberAuth) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
	}
}

// Annotations 定义会员认证信息表数据库注解
func (MemberAuth) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),               // 启用数据库字段注释
		schema.Comment("会员认证信息表"),               // 设置数据库表注释
		entsql.Annotation{Table: "member_auth"}, // 设置数据库表名
	}
}
