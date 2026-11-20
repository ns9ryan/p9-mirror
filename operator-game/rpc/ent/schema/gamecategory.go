package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GameCategory holds the schema definition for the GameCategory entity.
type GameCategory struct {
	ent.Schema
}

// Table of the GameCategory.
func (GameCategory) Table() string {
	return "game_category"
}

// Fields of the GameCategory.
func (GameCategory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("op_code"),
		field.String("category_code"),
		field.Int64("sort_no"),
		field.Int64("status").Comment("Status | 状态"),
		field.Time("created_at").Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the GameCategory.
func (GameCategory) Edges() []ent.Edge {
	return nil
}

// Indexes of the GameCategory.
func (GameCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("op_code", "category_code").Unique(),
	}
}

func (GameCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                 // 启用数据库字段注释
		schema.Comment("分站游戏分类表"),                 // 设置数据库表注释
		entsql.Annotation{Table: "game_category"}, // 设置数据库表名
	}
}
