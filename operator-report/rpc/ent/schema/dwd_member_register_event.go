// 由 ent 生成，请勿手动编辑。

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DwdMemberRegisterEvent 定义会员注册事件表结构
// 用途：记录新增活跃会员标记来自活动订单
// 数据粒度：按会员注册事件的粒度存储
type DwdMemberRegisterEvent struct {
	ent.Schema
}

// Fields 定义 DwdMemberRegisterEvent 的字段
func (DwdMemberRegisterEvent) Fields() []ent.Field {
	return []ent.Field{
		// ==================== 事件ID ====================
		// 设为主键ID
		field.String("id").
			NotEmpty().
			MaxLen(100).
			Comment("消息ID - 对应message_trail.sync_id"),

		// ==================== 维度列 ====================
		field.String("op_code").
			NotEmpty().
			MaxLen(50).
			Comment("分站编码"),

		field.String("agent_id").
			NotEmpty().
			MaxLen(50).
			Comment("代理ID"),

		field.String("user_id").
			NotEmpty().
			MaxLen(50).
			Comment("会员ID"),

		field.String("register_source").
			MaxLen(50).
			Optional().
			Comment("来源 (北美、非北美)"),

		field.Time("register_time").
			SchemaType(map[string]string{
				"mysql": "datetime",
				"doris": "datetime",
			}).
			Comment("注册时间"),

		// ==================== 会员属性 ====================
		field.Int8("is_activity").
			Default(0).
			Comment("是否参与活动"),

		// ==================== 元数据 ====================
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				"mysql": "datetime",
				"doris": "datetime",
			}).
			Comment("创建时间"),
	}
}

// Indexes 定义 DwdMemberRegisterEvent 的索引
func (DwdMemberRegisterEvent) Indexes() []ent.Index {
	return []ent.Index{
		// 综合查询索引
		index.Fields("op_code", "register_time"),
		index.Fields("op_code", "agent_id", "register_time"),
		index.Fields("op_code", "user_id", "register_time"),
		index.Fields("id"),
	}
}

// Annotations 返回表级别注解
func (DwdMemberRegisterEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "dwd_member_register_event",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("会员注册事件表 - 新增活跃会员标记来自活动订单"),
	}
}
