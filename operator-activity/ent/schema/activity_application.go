package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"
)

// ActivityApplication 定义活动申请表结构。
type ActivityApplication struct {
	ent.Schema
}

// Fields 定义活动申请表字段
func (ActivityApplication) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").
			Comment("申请的活动 ID"),

		field.Int64("member_id").
			Comment("申请会员 ID"),

		field.Int("apply_mode").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("申请方式，取值见业务常量"),

		field.Time("apply_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("申请时间"),

		field.String("apply_ip").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Comment("申请客户端 IP"),

		field.String("apply_fp").
			MaxLen(128).
			Default("").
			Comment("申请设备指纹/指纹哈希"),
	}
}

// Indexes 定义活动申请表索引
func (ActivityApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "activity_id", "member_id").
			Unique().
			StorageKey("uk_activity_application_activity_member"),
	}
}

// Mixin 定义活动申请表公共字段
func (ActivityApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动申请表查询隔离
func (ActivityApplication) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动申请表数据库注解
func (ActivityApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动申请表"),
		entsql.Annotation{
			Table: "activity_application",
		},
	}
}
