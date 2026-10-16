package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Provider 全局 IP 定位服务商。租户启用/优先级见 ProviderPolicy。
type Provider struct {
	ent.Schema
}

func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Immutable().
			Comment("主键"),
		field.String("name").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("服务商名称"),
		field.String("driver").
			MaxLen(32).
			NotEmpty().
			Comment("http / mmdb / ip2region / local"),
		field.String("tag").
			MaxLen(16).
			Default("").
			Comment("在线 / 本地 / 离线"),
		field.Bool("is_enabled").
			Default(true).
			Comment("默认是否参与查询；租户可被策略覆盖"),
		field.Int("priority").
			Default(100).
			Comment("默认优先级，越小越先；租户可被策略覆盖"),
		field.String("endpoint").
			MaxLen(512).
			Default("").
			Comment("URL 模板或离线库路径"),
		field.String("auth_type").
			MaxLen(16).
			Default("none").
			Comment("none / basic / bearer / header"),
		field.String("access_key").
			MaxLen(256).
			Default("").
			Comment("账号或自定义头名"),
		field.String("secret_key").
			MaxLen(256).
			Default("").
			Comment("密钥"),
		field.Int("timeout_sec").
			Default(5).
			Comment("HTTP 超时秒数"),
		field.JSON("extra_json", ExtraJSON{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("field_map / headers"),
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
		index.Fields("priority"),
	}
}

func (Provider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("IP geo provider"),
		entsql.Annotation{Table: "ip_geo_provider"},
	}
}
