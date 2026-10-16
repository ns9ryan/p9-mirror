package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberVipLevelRecord 定义会员VIP等级变更记录表结构
type MemberVipLevelRecord struct {
	ent.Schema
}

// Fields 定义会员VIP等级变更记录表字段
func (MemberVipLevelRecord) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 会员关联 ====================

		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		// ==================== 等级变化 ====================

		field.Int64("before_vip_level_id").
			Positive().
			Immutable().
			Comment("变更前VIP等级ID"),

		field.Int64("after_vip_level_id").
			Positive().
			Immutable().
			Comment("变更后VIP等级ID"),

		field.Int64("change_type").
			Range(1, 4).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Immutable().
			Comment("变化类型: 1系统晋级, 2系统降级, 3后台调整, 4规则重新计算"),

		// ==================== 变更信息 ====================

		field.String("change_reason").
			MaxLen(255).
			Optional().
			Nillable().
			Immutable().
			Comment("变化原因"),

		field.String("operator_user_code").
			MaxLen(64).
			Optional().
			Nillable().
			Immutable().
			Comment("后台操作账号"),

		// ==================== 时间 ====================

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("创建时间"),
	}
}

// Edges 定义会员VIP等级变更记录表关联关系
func (MemberVipLevelRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("vip_level_records").
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Mixin 定义会员VIP等级变更记录表公共字段
func (MemberVipLevelRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
	}
}

// Annotations 定义会员VIP等级变更记录表数据库注解
func (MemberVipLevelRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("会员VIP等级变更记录表"),
		entsql.Annotation{Table: "member_vip_level_record"},
	}
}
