package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// Member 定义会员主体表结构
type Member struct {
	ent.Schema
}

// Fields 定义会员主体表字段
func (Member) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 基础标识 ====================

		field.UUID("global_id", uuid.UUID{}).
			Unique().
			Immutable().
			Default(uuid.New).
			Comment("会员全局唯一标识"),

		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.String("code").
			NotEmpty().
			MinLen(12).
			MaxLen(12).
			Match(regexp.MustCompile(`^[1-9][0-9]{11}$`)).
			Immutable().
			Comment("会员对外编号, 12位数字"),

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

		field.String("phone").
			NotEmpty().
			MaxLen(16).
			Optional().
			Nillable().
			Comment("绑定手机号, 国际E.164格式"),

		field.String("email").
			NotEmpty().
			MaxLen(254).
			Optional().
			Nillable().
			Comment("绑定邮箱"),

		field.String("password_hash").
			NotEmpty().
			MaxLen(255).
			Optional().
			Nillable().
			Sensitive().
			Comment("登录密码哈希"),

		// ==================== 代理归属 ====================

		field.UUID("agent_global_id", uuid.UUID{}).
			Immutable().
			Comment("所属分层代理全局标识"),

		// ==================== 会员等级 ====================

		field.Int64("auto_level_id").
			Positive().
			Comment("当前自动层级ID"),

		field.Int64("vip_level_id").
			Positive().
			Comment("当前VIP等级ID"),

		// ==================== 会员状态 ====================

		field.Int64("status").
			Default(1).
			Range(1, 3).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("会员状态: 1正常, 2冻结, 3禁用"),

		// ==================== 注册信息 ====================

		field.String("register_method").
			NotEmpty().
			MaxLen(32).
			Immutable().
			Comment("注册方式"),

		field.String("register_channel").
			NotEmpty().
			MaxLen(32).
			Immutable().
			Comment("注册渠道"),

		field.String("register_ip").
			MaxLen(45).
			Optional().
			Nillable().
			Immutable().
			Comment("注册IP"),

		field.String("register_device_code").
			MaxLen(128).
			Optional().
			Nillable().
			Immutable().
			Comment("注册设备标识"),

		field.String("register_domain").
			MaxLen(255).
			Optional().
			Nillable().
			Immutable().
			Comment("注册域名"),

		// ==================== 最后登录信息 ====================

		field.Time("last_login_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("最后登录时间"),

		field.String("last_login_ip").
			MaxLen(45).
			Optional().
			Nillable().
			Comment("最后登录IP"),

		field.String("last_login_device_code").
			MaxLen(128).
			Optional().
			Nillable().
			Comment("最后登录设备标识"),
	}
}

// Edges 定义会员主体表关联关系
func (Member) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("auto_level", MemberAutoLevel.Type).
			Field("auto_level_id").
			Unique().
			Required(),

		edge.To("vip_level", MemberVipLevel.Type).
			Field("vip_level_id").
			Unique().
			Required(),

		edge.To("profile", MemberProfile.Type).
			Unique(),

		edge.To("oauth_identities", MemberOAuthIdentity.Type),

		edge.To("fixed_level_assignments", MemberFixedLevelAssignment.Type),

		edge.To("vip_progress", MemberVipProgress.Type).
			Unique(),

		edge.To("vip_level_records", MemberVipLevelRecord.Type),
	}
}

// Indexes 定义会员主体表索引
func (Member) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个厅的会员对外编号唯一
		index.Fields("operator_code", "code").
			Unique().
			StorageKey("uk_member_operator_code"),

		// 同一个厅的会员账号唯一, 不区分大小写
		index.Fields("operator_code", "username").
			Unique().
			StorageKey("uk_member_operator_username"),

		// 同一个厅的手机号唯一
		index.Fields("operator_code", "phone").
			Unique().
			StorageKey("uk_member_operator_phone"),

		// 同一个厅的邮箱唯一
		index.Fields("operator_code", "email").
			Unique().
			StorageKey("uk_member_operator_email"),

		// 查询指定分层代理下的会员
		index.Fields("operator_code", "agent_global_id").
			StorageKey("idx_member_operator_agent"),
	}
}

// Mixin 定义会员主体表公共字段
func (Member) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员主体表数据库注解
func (Member) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),          // 启用数据库字段注释
		schema.Comment("会员主体表"),            // 设置数据库表注释
		entsql.Annotation{Table: "member"}, // 设置数据库表名
	}
}
