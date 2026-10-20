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

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
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

		// ==================== 代理归属与身份 ====================

		field.Int64("agent_id").
			Positive().
			Immutable().
			Comment("会员所属分层代理ID"),

		field.Int64("agent_account_id").
			Positive().
			Optional().
			Nillable().
			Comment("会员自身的分层代理账号ID"),

		// ==================== 会员邀请 ====================

		field.Int64("inviter_member_id").
			Positive().
			Optional().
			Nillable().
			Immutable().
			Comment("邀请当前会员注册的会员ID"),

		field.String("invite_code").
			NotEmpty().
			MaxLen(64).
			Match(trimmedNonEmpty).
			Immutable().
			Comment("会员专属邀请码"),

		// ==================== 会员等级 ====================

		field.Int64("auto_level_id").
			Positive().
			Comment("当前自动层级ID"),

		field.Int64("vip_level_id").
			Positive().
			Comment("当前VIP等级ID"),

		// ==================== 会员状态与业务限制 ====================

		field.Int64("status").
			Default(1).
			Range(1, 3).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("会员状态: 1正常, 2冻结, 3禁用"),

		field.Bool("chat_disabled").
			Default(false).
			Comment("是否禁止发言"),

		field.Bool("betting_disabled").
			Default(false).
			Comment("是否禁止投注"),

		field.Bool("withdrawal_disabled").
			Default(false).
			Comment("是否禁止提款"),

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

		// ==================== 代理归属与身份 ====================

		edge.To("agent", AgentUser.Type).
			Field("agent_id").
			Unique().
			Required().
			Immutable(),

		edge.To("agent_account", AgentUser.Type).
			Field("agent_account_id").
			Unique(),

		// ==================== 会员邀请 ====================

		edge.To("invited_members", Member.Type),

		edge.From("inviter", Member.Type).
			Ref("invited_members").
			Field("inviter_member_id").
			Unique().
			Immutable(),

		// ==================== 会员等级 ====================

		edge.To("auto_level", MemberAutoLevel.Type).
			Field("auto_level_id").
			Unique().
			Required(),

		edge.To("vip_level", MemberVipLevel.Type).
			Field("vip_level_id").
			Unique().
			Required(),

		edge.To("fixed_level_assignments", MemberFixedLevelAssignment.Type),

		// ==================== 认证与个人资料 ====================

		edge.To("auth", MemberAuth.Type).
			Unique(),

		edge.To("profile", MemberProfile.Type).
			Unique(),

		edge.To("oauth_identities", MemberOAuthIdentity.Type),

		// ==================== VIP业务记录 ====================

		edge.To("vip_progress", MemberVipProgress.Type).
			Unique(),

		edge.To("vip_level_records", MemberVipLevelRecord.Type),
	}
}

// Indexes 定义会员主体表索引
func (Member) Indexes() []ent.Index {
	return []ent.Index{

		// ==================== 唯一约束 ====================

		// 同一个厅的会员对外编号唯一
		index.Fields("operator_code", "code").
			Unique().
			StorageKey("uk_member_operator_code"),

		// 同一个厅的会员邀请码唯一
		index.Fields("operator_code", "invite_code").
			Unique().
			StorageKey("uk_member_operator_invite_code"),

		// 一个代理账号最多绑定一个会员身份
		index.Fields("agent_account_id").
			Unique().
			StorageKey("uk_member_agent_account"),

		// ==================== 业务查询 ====================

		// 查询指定分层代理下的会员
		index.Fields("operator_code", "agent_id").
			StorageKey("idx_member_operator_agent"),

		// 查询指定会员邀请的下级会员
		index.Fields("operator_code", "inviter_member_id").
			StorageKey("idx_member_operator_inviter"),
	}
}

// Mixin 定义会员主体表公共字段
func (Member) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors 定义会员主体表查询隔离
func (Member) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks 定义会员主体表写入隔离
func (Member) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
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
