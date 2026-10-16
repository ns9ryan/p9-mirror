package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// IpGeoCache IP 查询 KV 缓存。换 Redis 时仍是同一套 key + JSON。
type IpGeoCache struct {
	ent.Schema
}

func (IpGeoCache) Fields() []ent.Field {
	jsonb := map[string]string{dialect.Postgres: "jsonb"}
	return []ent.Field{
		field.String("id").
			StorageKey("ip").
			MaxLen(160).
			NotEmpty().
			Immutable().
			Comment("缓存键 name|规范化 IP"),
		field.JSON("payload", json.RawMessage{}).
			SchemaType(jsonb).
			Comment("Record JSON 整包"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("过期时间；空表示不过期。表驱动 Touch 只改这一列"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

func (IpGeoCache) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("IP geo KV cache"),
		entsql.Annotation{Table: "ip_geo_cache"},
	}
}
