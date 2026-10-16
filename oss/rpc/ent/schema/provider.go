package schema

import (
	"time"

	"oa.98ent.com/p9/core/common/entmixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Provider 一份存储驱动（连接 + 桶）。按 JWT operator_code 隔离；启用后覆盖本租户 yaml/etcd 的 oss 段。
type Provider struct {
	ent.Schema
}

func (Provider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.OperatorCodeMixin{},
	}
}

func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Positive().
			Immutable().
			Comment("主键"),
		field.String("name").
			MaxLen(64).
			NotEmpty().
			Comment("服务商名称"),
		field.Bool("is_default").
			Default(false).
			Comment("是否为当前启用的默认配置"),
		field.String("driver").
			MaxLen(32).
			Default("s3").
			Comment("存储驱动"),
		field.String("endpoint").
			MaxLen(512).
			Default("").
			Comment("S3 API 地址"),
		field.String("file_url").
			MaxLen(512).
			Default("").
			Comment("公开读 URL 前缀"),
		field.String("access_key").
			MaxLen(256).
			Default("").
			Comment("访问密钥"),
		field.String("secret_key").
			MaxLen(256).
			Default("").
			Comment("密钥"),
		field.Bool("path_style").
			Default(true).
			Comment("S3 path-style"),
		field.String("region").
			MaxLen(64).
			Default("").
			Comment("region"),
		field.String("buckets_json").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("oss.buckets JSON"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

func (Provider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "name").Unique(),
		index.Fields("operator_code", "is_default"),
	}
}

func (Provider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("OSS storage provider"),
		entsql.Annotation{Table: "oss_provider"},
	}
}
