package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
)

// EventInbox 定义 Kafka 事件幂等消费记录表结构
type EventInbox struct {
	ent.Schema
}

// Fields 定义事件 Inbox 字段
func (EventInbox) Fields() []ent.Field {
	return []ent.Field{
		field.String("kafka_id").
			MaxLen(128).
			NotEmpty().
			Comment("Kafka 消息 ID"),

		field.String("msg_value").
			Default("").
			Comment("Kafka 消息原始值"),

		field.String("group").
			MaxLen(128).
			NotEmpty().
			Comment("Kafka Consumer Group"),

		field.String("topic").
			MaxLen(128).
			NotEmpty().
			Comment("Kafka Topic"),

		field.Int64("offset").
			Default(0).
			Comment("偏移量"),

		field.Int64("partition").
			Default(0).
			Comment("分区"),
	}
}

// Indexes 定义事件 Inbox 表索引
func (EventInbox) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group", "topic", "partition", "offset").
			Unique().
			StorageKey("idx_event_inbox_unique"),
	}
}

// Mixin 定义事件 Inbox 表公共字段
func (EventInbox) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// // Interceptors 定义事件 Inbox 表查询隔离
// func (EventInbox) Interceptors() []ent.Interceptor {
// 	return []ent.Interceptor{
// 		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
// 			entmixin.FilterOperatorCode(ctx, q)
// 			return nil
// 		}),
// 	}
// }

// Annotations 定义事件 Inbox 表数据库注解
func (EventInbox) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Kafka 事件幂等消费记录表"),
		entsql.Annotation{
			Table: "event_inbox",
		},
	}
}
