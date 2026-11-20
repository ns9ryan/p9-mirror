package schema

import (
	"encoding/json"

	"oa.98ent.com/p9/node-dispatch/rpc/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DispatchTask 定义调度任务表结构
type DispatchTask struct {
	ent.Schema
}

// Fields 定义调度任务表字段
func (DispatchTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("task_no").
			NotEmpty().
			MaxLen(64).
			Unique().
			Immutable().
			Comment("调度中心生成的全局唯一任务编号"),

		field.String("request_no").
			NotEmpty().
			MaxLen(64).
			Unique().
			Immutable().
			Comment("调用方生成的请求编号"),

		field.String("target").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("任务目标服务"),

		field.String("task_type").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("任务类型"),

		field.JSON("params", json.RawMessage{}).
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("任务执行所需的最小参数"),

		field.Int64("node_id").
			Immutable().
			Comment("执行节点本地主键"),

		field.Int64("status").
			Default(1).
			Range(1, 4).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("任务状态: 1待执行, 2执行中, 3成功, 4失败"),

		field.JSON("result", json.RawMessage{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("任务执行结果"),

		field.String("error_message").
			MaxLen(2000).
			Optional().
			Nillable().
			Comment("任务执行失败原因"),

		field.Time("started_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("开始执行时间"),

		field.Time("finished_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("执行结束时间"),
	}
}

// Edges 定义调度任务表关联关系
func (DispatchTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("node", Node.Type).
			Ref("dispatch_tasks").
			Field("node_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义调度任务表索引
func (DispatchTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("node_id"),
	}
}

// Mixin 定义调度任务表公共字段
func (DispatchTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Annotations 定义调度任务表数据库注解
func (DispatchTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                 // 启用数据库字段注释
		schema.Comment("调度任务表"),                   // 设置数据库表注释
		entsql.Annotation{Table: "dispatch_task"}, // 设置数据库表名
	}
}
