package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
)

// TaskConfig 定义分站任务配置表结构
type TaskConfig struct {
	ent.Schema
}

// Fields 定义任务配置字段
func (TaskConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("task_code").
			MaxLen(64).
			NotEmpty().
			Comment("任务编码，如 REGISTER、DAILY_DEPOSIT 等"),

		field.String("category").
			MaxLen(16).
			NotEmpty().
			Comment("任务分类：ONCE|DAILY|WEEKLY|MONTHLY"),

		field.String("name").
			MaxLen(255).
			NotEmpty().
			Comment("任务名称"),

		field.String("name_i18n").
			Default("{}").
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("任务名称国际化 JSON"),

		field.String("description").
			MaxLen(1000).
			Default("").
			Comment("任务描述"),

		field.String("description_i18n").
			Default("{}").
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("任务描述国际化 JSON"),

		field.String("timezone").
			MaxLen(64).
			Default("UTC").
			Comment("分站时区，如 Asia/Shanghai"),

		field.Int("status").
			Default(2).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("任务状态：1=启用，2=禁用"),

		field.Int("completion_amount").
			Default(0).
			NonNegative().
			Comment("完成条件USD金额"),

		// field.Int("reward_rank").
		// 	Default(1).
		// 	SchemaType(map[string]string{
		// 		dialect.Postgres: "smallint",
		// 	}).
		// 	Comment("奖励等级"),

		field.Int64("reward_amount").
			Default(0).
			NonNegative().
			Comment("奖励金额（最小货币单位）"),

		// field.String("reward_currency").
		// 	MaxLen(16).
		// 	Optional().
		// 	Comment("奖励货币代码"),

		field.String("payout_mode").
			MaxLen(16).
			Default("MANUAL").
			Comment("派发方式：MANUAL|AUTO"),

		// field.Int("config_version").
		// 	Default(1).
		// 	NonNegative().
		// 	SchemaType(map[string]string{
		// 		dialect.Postgres: "integer",
		// 	}).
		// 	Comment("配置版本号"),

		field.Int64("sort").
			Default(0).
			Comment("任务排序"),

		// 放到报表服务统计
		// field.Int64("claimed_count").
		// 	Default(0).
		// 	NonNegative().
		// 	Comment("已领取计数（缓存字段）"),

		// field.Int64("claimed_amount").
		// 	Default(0).
		// 	NonNegative().
		// 	Comment("已领取总额（缓存字段）"),

		// field.Time("effective_at").
		// 	SchemaType(map[string]string{
		// 		dialect.Postgres: "timestamptz(3)",
		// 	}).
		// 	Comment("生效时间, 可为空，用于指定任务生效的起始时间"),
	}
}

// Indexes 定义任务配置表索引
func (TaskConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "task_code").
			Unique().
			StorageKey("idx_task_config_unique"),
		index.Fields("operator_code", "task_code", "status").
			StorageKey("idx_task_config_operator_code_task_code_status"),
	}
}

// Mixin 定义任务配置表公共字段
func (TaskConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// // Interceptors 定义任务配置表查询隔离
// func (TaskConfig) Interceptors() []ent.Interceptor {
// 	return []ent.Interceptor{
// 		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
// 			entmixin.FilterOperatorCode(ctx, q)
// 			return nil
// 		}),
// 	}
// }

// Annotations 定义任务配置表数据库注解
func (TaskConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("分站任务配置表"),
		entsql.Annotation{
			Table: "task_config",
		},
	}
}
