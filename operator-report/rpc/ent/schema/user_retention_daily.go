// 由 ent 生成，请勿手动编辑。

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserRetentionDaily 定义 mv_user_retention_daily 物化视图结构
// 用途：用户留存分析、按注册日期维度的新用户留存统计
// 数据粒度：按 op_code + agent_id + register_date 维度聚合的新用户留存数据
type UserRetentionDaily struct {
	ent.Schema
}

// Fields 定义 UserRetentionDaily 的字段
func (UserRetentionDaily) Fields() []ent.Field {
	return []ent.Field{
		// ==================== 维度列 ====================
		field.String("op_code").
			NotEmpty().
			MaxLen(64).
			Comment("维度：分站编码"),

		field.Int64("agent_id").
			Comment("维度：代理ID"),

		field.Time("register_date").
			SchemaType(map[string]string{
				"mysql": "date",
				"doris": "date",
			}).
			Comment("维度：用户注册日期（YYYY-MM-DD）"),

		// ==================== 新用户统计 ====================
		field.Int64("total_new_users").
			Default(0).
			Comment("新用户总数"),

		// ==================== 1日留存 ====================
		field.Int64("retention_1d_users").
			Default(0).
			Comment("注册后第1天至少登录一次的用户数"),

		// ==================== 2日留存 ====================
		field.Int64("retention_2d_users").
			Default(0).
			Comment("注册后第1～2天至少登录一次的用户数"),

		// ==================== 3日留存 ====================
		field.Int64("retention_3d_users").
			Default(0).
			Comment("注册后第1～3天至少登录一次的用户数"),

		// ==================== 7日留存 ====================
		field.Int64("retention_7d_users").
			Default(0).
			Comment("注册后第1～7天至少登录一次的用户数"),

		// ==================== 15日留存 ====================
		field.Int64("retention_15d_users").
			Default(0).
			Comment("注册后第1～15天至少登录一次的用户数"),

		// ==================== 30日留存 ====================
		field.Int64("retention_30d_users").
			Default(0).
			Comment("注册后第1～30天至少登录一次的用户数"),

		// ==================== 60日留存 ====================
		field.Int64("retention_60d_users").
			Default(0).
			Comment("注册后第1～60天至少登录一次的用户数"),
	}
}

// Indexes 定义 UserRetentionDaily 的索引
func (UserRetentionDaily) Indexes() []ent.Index {
	return []ent.Index{
		// 联合主键：唯一标识一条留存记录
		index.Fields("op_code", "agent_id", "register_date").Unique(),
		// 分站查询索引
		index.Fields("op_code", "agent_id"),
		// 分站和注册日期联合查询
		index.Fields("op_code", "register_date"),
		// 注册日期查询
		index.Fields("register_date"),
	}
}

func (UserRetentionDaily) Table() string {
	return "mv_user_retention_daily"
}

func (UserRetentionDaily) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户留存分析物化视图，统计新用户按注册日期的多时段留存情况（1日、2日、3日、7日、15日、30日、60日）"),
		entsql.Annotation{Table: "mv_user_retention_daily"},
	}
}

// Mixin 禁用默认的自增主键
func (UserRetentionDaily) Mixin() []ent.Mixin {
	return []ent.Mixin{}
}
