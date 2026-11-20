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

// DwdMemberLoginEvent 定义会员登录事件表结构
// 用途：存储用户登录事件，支持通过SQL动态计算留存统计
// 数据粒度：按会员登录事件的粒度存储
type DwdMemberLoginEvent struct {
	ent.Schema
}

// Fields 定义 DwdMemberLoginEvent 的字段
func (DwdMemberLoginEvent) Fields() []ent.Field {
	return []ent.Field{
		// ==================== 事件ID ====================
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

		field.Time("login_time").
			SchemaType(map[string]string{
				"mysql": "datetime",
				"doris": "datetime",
			}).
			Comment("登录时间"),

		field.String("login_source").
			MaxLen(50).
			Optional().
			Comment("登录来源 (北美、非北美)"),

		field.Time("register_time").
			SchemaType(map[string]string{
				"mysql": "datetime",
				"doris": "datetime",
			}).
			Comment("注册时间"),

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

// Indexes 定义 DwdMemberLoginEvent 的索引
func (DwdMemberLoginEvent) Indexes() []ent.Index {
	return []ent.Index{
		// 综合查询索引
		index.Fields("op_code", "login_time"),
		index.Fields("op_code", "agent_id", "login_time"),
		index.Fields("op_code", "user_id", "login_time"),
		index.Fields("op_code", "register_time", "login_time"),
		index.Fields("id"),
	}
}

// Annotations 返回表级别注解
func (DwdMemberLoginEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "dwd_member_login_event",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("登录事件明细表 - 存储用户登录事件，支持通过SQL动态计算留存统计"),
	}
}
