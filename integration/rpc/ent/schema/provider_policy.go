package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProviderPolicy 租户对某个全局服务商的启用/优先级。没有行则用 Provider 默认值。
type ProviderPolicy struct {
	ent.Schema
}

func (ProviderPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Positive().
			Immutable().
			Comment("主键"),
		field.String("operator_code").
			MaxLen(64).
			NotEmpty().
			Comment("租户编码"),
		field.Int64("provider_id").
			Positive().
			Comment("ip_geo_provider.id"),
		field.Bool("is_enabled").
			Default(true).
			Comment("该租户是否启用"),
		field.Int("priority").
			Default(100).
			Comment("该租户优先级，越小越先"),
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

func (ProviderPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "provider_id").Unique(),
		index.Fields("operator_code"),
	}
}

func (ProviderPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("IP geo provider policy per tenant"),
		entsql.Annotation{Table: "ip_geo_provider_policy"},
	}
}
