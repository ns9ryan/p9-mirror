package schema

import (
	"time"

	"oa.98ent.com/p9/core/common/entmixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// File 上传档案。业务表仍存 object_key。多租户按 JWT operator_code 隔离。
type File struct {
	ent.Schema
}

func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.OperatorCodeMixin{},
	}
}

func (File) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Positive().
			Immutable().
			Comment("主键"),
		field.String("object_key").
			MaxLen(512).
			NotEmpty().
			Comment("对象 key"),
		field.String("scene").
			MaxLen(32).
			NotEmpty().
			Comment("上传场景"),
		field.String("bucket").
			MaxLen(128).
			NotEmpty().
			Comment("所在桶"),
		field.String("name").
			MaxLen(255).
			Default("").
			Comment("原始文件名"),
		field.Int64("size").
			NonNegative().
			Comment("字节数"),
		field.String("content_type").
			MaxLen(128).
			Default("").
			Comment("Content-Type"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("object_key").Unique(),
		index.Fields("operator_code"),
		index.Fields("scene"),
		index.Fields("created_at"),
	}
}

func (File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("OSS file archive"),
		entsql.Annotation{Table: "oss_file"},
	}
}
