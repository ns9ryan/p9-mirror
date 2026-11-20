package schema

import (
	"time"

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

// MemberVipProgress 定义会员VIP晋级进度表结构
type MemberVipProgress struct {
	ent.Schema
}

// Fields 定义会员VIP晋级进度表字段
func (MemberVipProgress) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 会员关联 ====================

		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		// ==================== 晋级进度 ====================

		field.Int64("upgrade_deposit").
			NonNegative().
			Default(0).
			Comment("当前VIP等级已完成的存款晋级进度"),

		field.Int64("upgrade_water").
			NonNegative().
			Default(0).
			Comment("当前VIP等级已完成的洗码晋级进度"),

		// ==================== 进度时间 ====================

		field.Time("progress_started_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("当前VIP晋级周期开始时间"),
	}
}

// Edges 定义会员VIP晋级进度表关联关系
func (MemberVipProgress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("vip_progress").
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员VIP晋级进度表索引
func (MemberVipProgress) Indexes() []ent.Index {
	return []ent.Index{
		// 一个会员只有一条VIP晋级进度
		index.Fields("member_id").
			Unique().
			StorageKey("uk_member_vip_progress_member"),

		// 按厅编码查询晋级进度
		index.Fields("operator_code").
			StorageKey("idx_member_vip_progress_operator_code"),
	}
}

// Mixin 定义会员VIP晋级进度表公共字段
func (MemberVipProgress) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors VIP晋级进度查询隔离
func (MemberVipProgress) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks VIP晋级进度写入隔离
func (MemberVipProgress) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
	}
}

// Annotations 定义会员VIP晋级进度表数据库注解
func (MemberVipProgress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),    // 启用数据库字段注释
		schema.Comment("会员VIP晋级进度表"), // 设置数据库表注释
		entsql.Annotation{Table: "member_vip_progress"},
	}
}
