package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GameChannel holds the schema definition for the GameChannel entity.
type GameChannel struct {
	ent.Schema
}

// Table of the GameChannel.
func (GameChannel) Table() string {
	return "game_channel"
}

// Fields of the GameChannel.
func (GameChannel) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("op_code"),
		field.String("channel_code"),
		field.Int64("sort_no"),
		field.Int64("load_type"),
		field.Int64("status"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the GameChannel.
func (GameChannel) Edges() []ent.Edge {
	return nil
}

// Indexes of the GameChannel.
func (GameChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("op_code", "channel_code").Unique(),
	}
}

func (GameChannel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                // 启用数据库字段注释
		schema.Comment("分站游戏渠道表"),                // 设置数据库表注释
		entsql.Annotation{Table: "game_channel"}, // 设置数据库表名
	}
}
