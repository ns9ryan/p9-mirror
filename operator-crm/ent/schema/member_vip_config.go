package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberVipConfig 定义会员VIP配置表结构
type MemberVipConfig struct {
	ent.Schema
}

// Fields 定义会员VIP配置表字段
func (MemberVipConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Unique().
			Immutable().
			Comment("所属厅业务编码"),

		field.Bool("vip_enabled").
			Default(false).
			Comment("VIP总开关"),

		field.Bool("upgrade_reward_enabled").
			Default(false).
			Comment("晋级奖励开关"),

		field.Int64("reward_water_multiplier").
			NonNegative().
			Default(100000).
			Comment("VIP晋级奖励洗码倍率，放大100000倍存储"),

		field.Bool("retention_check_enabled").
			Default(false).
			Comment("保级考核开关"),

		field.Int64("reward_grant_mode").
			Default(1).
			Range(1, 2).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("奖励派发模式: 1手动领取, 2自动派发"),

		field.Time("last_enabled_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("VIP最近一次开启时间"),
	}
}

// Indexes 定义会员VIP配置表索引
func (MemberVipConfig) Indexes() []ent.Index {
	return nil
}

// Mixin 定义会员VIP配置表公共字段
func (MemberVipConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义会员VIP配置表数据库注解
func (MemberVipConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                     // 启用数据库字段注释
		schema.Comment("会员VIP配置表"),                    // 设置数据库表注释
		entsql.Annotation{Table: "member_vip_config"}, // 设置数据库表名
	}
}
