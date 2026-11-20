package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/operator-base/rpc/ent/schema/mixins"
)

// OperatorSiteSetting 保存各分站独立的站点设置，不保存登录次数计数器。
type OperatorSiteSetting struct {
	ent.Schema
}

// Fields 定义站点标题、Logo、登录阈值及修改人字段。
func (OperatorSiteSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("operator_id").
			Unique().
			Immutable().
			Comment("所属站点本地主键，每站仅一条配置"),

		field.String("site_title").
			NotEmpty().
			MaxLen(100).
			Comment("站点标题"),

		field.String("site_logo_url").
			Default("").
			MaxLen(1024).
			Comment("站点 Logo 地址，空字符串表示未设置"),

		field.Int64("email_login_limit").
			Default(0).
			Range(0, 100).
			Comment("邮箱登录阈值，范围 0～100，默认 0 表示不限制；当前仅保存配置"),

		field.Int64("sms_login_limit").
			Default(0).
			Range(0, 100).
			Comment("短信登录阈值，范围 0～100，默认 0 表示不限制；当前仅保存配置"),

		field.Int64("member_password_error_limit").
			Default(0).
			Range(0, 100).
			Comment("会员密码错误次数上限，范围 0～100，默认 0 表示不限制"),

		field.Int64("ip_member_register_limit").
			Default(0).
			Range(0, 100).
			Comment("同一 IP 最多注册会员数，范围 0～100，默认 0 表示不限制"),

		field.Int64("register_captcha_mode").
			Default(0).
			Range(0, 2).
			Comment("注册页智能验证，默认 0 不验证，1 拖拽验证，2 图形验证"),

		field.Int64("login_captcha_mode").
			Default(0).
			Range(0, 2).
			Comment("登录页智能验证，默认 0 不验证，1 拖拽验证，2 图形验证"),

		field.Int64("updated_by").
			Default(0).
			NonNegative().
			Comment("最后修改人的登录用户 ID，由后端身份取得；默认 0 表示系统初始化"),
	}
}

// Edges 定义站点设置与所属 operator 的关联关系。
func (OperatorSiteSetting) Edges() []ent.Edge {
	return []ent.Edge{
		// 使用单向关联，无需在原有 Operator schema 中新增反向关联。
		edge.To("operator", Operator.Type).
			Field("operator_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义站点设置的唯一索引。
func (OperatorSiteSetting) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个 operator 只保存一条站点设置。
		index.Fields("operator_id").
			Unique().
			StorageKey("uk_operator_site_setting_operator"),
	}
}

// Mixin 复用项目现有的主键、创建时间和更新时间字段。
func (OperatorSiteSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义表名、数据库注释和阈值非负约束。
func (OperatorSiteSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("分站站点设置，每站独立保存一条记录"),
		entsql.Annotation{
			Table: "operator_site_setting",
			Checks: map[string]string{
				"operator_site_setting_email_limit_range":           "email_login_limit BETWEEN 0 AND 100",
				"operator_site_setting_sms_limit_range":             "sms_login_limit BETWEEN 0 AND 100",
				"operator_site_setting_password_error_limit_range":  "member_password_error_limit BETWEEN 0 AND 100",
				"operator_site_setting_ip_register_limit_range":     "ip_member_register_limit BETWEEN 0 AND 100",
				"operator_site_setting_register_captcha_mode_range": "register_captcha_mode BETWEEN 0 AND 2",
				"operator_site_setting_login_captcha_mode_range":    "login_captcha_mode BETWEEN 0 AND 2",
				"operator_site_setting_updated_by_nonnegative":      "updated_by >= 0",
			},
		},
	}
}
