package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GameOrder holds the schema definition for the GameOrder entity.
type GameOrder struct {
	ent.Schema
}

// Table of the GameOrder.
func (GameOrder) Table() string {
	return "game_order"
}

// Fields of the GameOrder.
func (GameOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("主键ID"),
		field.String("operator_code").Comment("分站编码"),
		field.Int64("member_id").Comment("会员ID"),
		field.Int64("amount").Comment("订单金额"),
		field.Int64("settle_at").Comment("结算时间(毫秒时间戳)"),
	}
}

// Edges of the GameOrder.
func (GameOrder) Edges() []ent.Edge {
	return nil
}

// Indexes of the GameOrder.
func (GameOrder) Indexes() []ent.Index {
	return []ent.Index{
		// 定时查询用：查询指定时间范围内的订单
		index.Fields("operator_code", "settle_at"),
		// 按会员查询优化
		index.Fields("operator_code", "member_id"),
	}
}

func (GameOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),              // 启用数据库字段注释
		schema.Comment("游戏订单表"),               // 设置数据库表注释
		entsql.Annotation{Table: "game_order"}, // 设置数据库表名
	}
}
