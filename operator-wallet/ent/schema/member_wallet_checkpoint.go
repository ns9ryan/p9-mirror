package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
)

// MemberWalletCheckpoint 定义钱包 Kafka 消费进度表
type MemberWalletCheckpoint struct {
	ent.Schema
}

// Fields 定义钱包 Kafka 消费进度字段
func (MemberWalletCheckpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("topic").
			NotEmpty().
			MaxLen(249).
			Immutable().
			Comment("Kafka 资金事件 Topic"),

		field.Int32("partition").
			NonNegative().
			Immutable().
			Comment("Kafka Partition 编号"),

		field.Int64("next_offset").
			NonNegative().
			Comment("下一条需要持久化的 Kafka Offset"),
	}
}

// Indexes 定义钱包 Kafka 消费进度索引
func (MemberWalletCheckpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("topic", "partition").
			Unique().
			StorageKey("uk_member_wallet_checkpoint_topic_partition"),
	}
}

// Mixin 定义钱包 Kafka 消费进度公共字段
func (MemberWalletCheckpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
	}
}

// Annotations 定义钱包 Kafka 消费进度数据库注解
func (MemberWalletCheckpoint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("会员钱包 Kafka 消费进度表"),
		entsql.Annotation{
			Table: "member_wallet_checkpoint",
			Checks: map[string]string{
				"ck_member_wallet_checkpoint_partition":   "partition >= 0",
				"ck_member_wallet_checkpoint_next_offset": "next_offset >= 0",
			},
		},
	}
}
