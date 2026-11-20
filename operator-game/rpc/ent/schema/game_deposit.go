package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GameDeposit holds the schema definition for the GameDeposit entity.
type GameDeposit struct {
	ent.Schema
}

// Table of the GameDeposit.
func (GameDeposit) Table() string {
	return "game_deposit"
}

// Fields of the GameDeposit.
func (GameDeposit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("主键ID"),
		field.String("operator_code").Comment("分站编码"),
		field.Int64("member_id").Comment("会员ID"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the GameDeposit.
func (GameDeposit) Edges() []ent.Edge {
	return nil
}

// Indexes of the GameDeposit.
func (GameDeposit) Indexes() []ent.Index {
	return []ent.Index{
		// 定时查询用：查询指定时间范围内的充值
		index.Fields("operator_code", "updated_at"),
		// 按会员查询优化
		index.Fields("operator_code", "member_id"),
	}
}

func (GameDeposit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                // 启用数据库字段注释
		schema.Comment("充值记录表"),                  // 设置数据库表注释
		entsql.Annotation{Table: "game_deposit"}, // 设置数据库表名
	}
}
