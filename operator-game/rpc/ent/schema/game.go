package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Game holds the schema definition for the Game entity.
type Game struct {
	ent.Schema
}

// Table of the Game.
func (Game) Table() string {
	return "game"
}

type CurrencyInfo struct {
	Code string `json:"code"`
}

// Fields of the Game.
func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("op_code"),
		field.String("game_code"),
		field.Int64("source_id").Optional(),
		field.String("category_code").Optional(),
		field.String("provider_code").Optional(),
		field.String("channel_code").Optional(),
		field.JSON("currency_list", []CurrencyInfo{}).Optional(),
		field.String("provider_key").Optional(),
		field.String("name").Optional(),
		field.String("image_url").Optional(),
		field.Int64("sort_no"),
		field.Bool("supports_embed"),
		field.Bool("supports_redirect"),
		field.Int64("status").Comment("Status | 状态"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the Game.
func (Game) Edges() []ent.Edge {
	return nil
}

// Indexes of the Game.
func (Game) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("op_code", "game_code").Unique(),
	}
}

func (Game) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),        // 启用数据库字段注释
		schema.Comment("分站游戏表"),          // 设置数据库表注释
		entsql.Annotation{Table: "game"}, // 设置数据库表名
	}
}
