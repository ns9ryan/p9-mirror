package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LocalIP 本地维护的 IP 地理信息，作为 driver=local 的数据源。
type LocalIP struct {
	ent.Schema
}

func (LocalIP) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Immutable().
			Comment("主键"),
		field.String("ip").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("规范化 IPv4 / IPv6"),
		field.String("continent").
			MaxLen(64).
			Default("").
			Comment("大洲名"),
		field.String("continent_code").
			MaxLen(8).
			Default("").
			Comment("大洲码"),
		field.String("country").
			MaxLen(64).
			Default("").
			Comment("国家名"),
		field.String("country_code").
			MaxLen(8).
			Default("").
			Comment("国家码"),
		field.String("province").
			MaxLen(64).
			Default("").
			Comment("省名"),
		field.String("province_iso_code").
			MaxLen(16).
			Default("").
			Comment("省码"),
		field.String("city").
			MaxLen(64).
			Default("").
			Comment("城市名"),
		field.String("postal").
			MaxLen(32).
			Default("").
			Comment("邮编"),
		field.Float("lat").
			Default(0).
			Comment("纬度"),
		field.Float("lon").
			Default(0).
			Comment("经度"),
		field.String("timezone").
			MaxLen(64).
			Default("").
			Comment("时区"),
		field.String("isp").
			MaxLen(255).
			Default("").
			Comment("ISP"),
		field.String("org").
			MaxLen(255).
			Default("").
			Comment("组织"),
		field.String("asn").
			MaxLen(255).
			Default("").
			Comment("ASN"),
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

func (LocalIP) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("country_code"),
	}
}

func (LocalIP) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("IP geo local table"),
		entsql.Annotation{Table: "ip_geo_local"},
	}
}
