package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"
)

// ActivityStatsClaimer 定义活动领取人去重统计表结构。
type ActivityStatsClaimer struct {
	ent.Schema
}

// Fields 定义活动领取人去重表字段
func (ActivityStatsClaimer) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("activity_id").
			Comment("活动 ID"),

		field.Int64("member_id").
			Comment("已领取过该活动的会员 ID"),
	}
}

// Indexes 定义活动领取人去重表索引
func (ActivityStatsClaimer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("activity_id", "member_id").
			Unique().
			StorageKey("uk_activity_stats_claimer_activity_member"),
	}
}

// Mixin 定义活动领取人去重表公共字段
func (ActivityStatsClaimer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动领取人去重表查询隔离
func (ActivityStatsClaimer) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动领取人去重表数据库注解
func (ActivityStatsClaimer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动领取人去重表"),
		entsql.Annotation{
			Table: "activity_stats_claimer",
		},
	}
}
