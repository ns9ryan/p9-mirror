package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GameChatLog holds the schema definition for the GameChatLog entity.
type GameChatLog struct {
	ent.Schema
}

// Table of the GameChatLog.
func (GameChatLog) Table() string {
	return "game_chat_log"
}

// Fields of the GameChatLog.
func (GameChatLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("主键ID"),
		field.String("operator_code").Comment("分站编码"),
		field.Int64("member_id").Comment("会员ID"),
		field.String("message").Comment("聊天消息内容"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the GameChatLog.
func (GameChatLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the GameChatLog.
func (GameChatLog) Indexes() []ent.Index {
	return []ent.Index{
		// 定时查询用：查询指定时间范围内的聊天消息
		index.Fields("operator_code", "updated_at"),
		// 按会员查询优化
		index.Fields("operator_code", "member_id"),
		// 按会员和时间范围查询优化
		index.Fields("operator_code", "member_id", "updated_at"),
	}
}

func (GameChatLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                 // 启用数据库字段注释
		schema.Comment("聊天日志表"),                  // 设置数据库表注释
		entsql.Annotation{Table: "game_chat_log"}, // 设置数据库表名
	}
}
