package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GameWithdraw holds the schema definition for the GameWithdraw entity.
type GameWithdraw struct {
	ent.Schema
}

// Table of the GameWithdraw.
func (GameWithdraw) Table() string {
	return "game_withdraw"
}

// Fields of the GameWithdraw.
func (GameWithdraw) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("主键ID"),
		field.String("operator_code").Comment("分站编码"),
		field.Int64("member_id").Comment("会员ID"),
		field.Time("updated_at").Default(time.Now).Comment("更新时间"),
	}
}

// Edges of the GameWithdraw.
func (GameWithdraw) Edges() []ent.Edge {
	return nil
}

// Indexes of the GameWithdraw.
func (GameWithdraw) Indexes() []ent.Index {
	return []ent.Index{
		// 定时查询用：查询指定时间范围内的提现
		index.Fields("operator_code", "updated_at"),
		// 按会员查询优化
		index.Fields("operator_code", "member_id"),
	}
}

func (GameWithdraw) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                  // 启用数据库字段注释
		schema.Comment("提现记录表"),                   // 设置数据库表注释
		entsql.Annotation{Table: "game_withdraw"}, // 设置数据库表名
	}
}
