package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// GameProvider holds the schema definition for the GameProvider entity.
type GameProvider struct {
	ent.Schema
}

// Table of the GameProvider.
func (GameProvider) Table() string {
	return "game_provider"
}

// Fields of the GameProvider.
func (GameProvider) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("provider_code"),
		field.String("channel_code").Optional(),
		field.String("logo_url").Optional(),
		field.Int64("sort_no"),
		field.Int64("status"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the GameProvider.
func (GameProvider) Edges() []ent.Edge {
	return nil
}

func (GameProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                 // 启用数据库字段注释
		schema.Comment("分站游戏供应商表"),                // 设置数据库表注释
		entsql.Annotation{Table: "game_provider"}, // 设置数据库表名
	}
}
