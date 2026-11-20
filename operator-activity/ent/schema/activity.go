package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-activity/ent/intercept"
)

// Activity 定义活动配置主表结构。
type Activity struct {
	ent.Schema
}

// Fields 定义活动表字段
func (Activity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("type").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动类型: 1-充值 2-洗码 3-推荐 4-幸运转盘..."),

		field.Int("status").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动状态: 1-草稿 2-进行中 3-已结束"),

		field.Int("apply_mode").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动参与模式: 1-自动参与 2-手动参与"),

		field.Int("cycle_mode").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动周期模式: 0-未指定 1-不循环 2-每日 3-每周 4-每月 5-无限循环"),

		field.Bool("is_permanent").
			Default(false).
			Comment("活动是否永久，为 true 时可不依赖结束时间"),

		field.Time("open_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("活动开始时间"),

		field.Time("close_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("活动结束时间"),

		field.Time("release_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("活动发布时间，未发布为空"),

		field.Bool("auto_audit").
			Default(false).
			Comment("是否开启自动审核发奖"),

		field.Int64("auto_audit_max_amt").
			Default(0).
			NonNegative().
			Comment("自动审核金额上限（最小货币单位），超过转人工"),

		field.Int("wagering_multiplier").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("洗码倍数，如 1 表示 1 倍"),

		field.Int("wagering_base").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("洗码计算基数: 1-本金+赠送 2-仅赠送"),

		field.Int("payout_mode").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("派发奖励模式: 1-手动领取过期作废 2-手动领取过期自动领 3-自动派发"),

		field.Int("payout_timing").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("派发奖励时间: 1-实时 2-次日"),

		field.Int("section").
			Default(1).
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Comment("活动所属: 1-大厅 2-电子 3-体育"),

		field.Int64("claim_expire_sec").
			Default(0).
			NonNegative().
			Comment("手动领取过期秒数，0 表示不过期"),

		field.String("settings").
			Default("{}").
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("活动配置 JSON"),

		field.JSON("allowed_clients", []int64{}).
			Default([]int64{}).
			Comment("允许参与的客户端列表，取值见业务常量，空表示不限制"),

		field.JSON("allowed_fixed_level", []int64{}).
			Default([]int64{}).
			Comment("允许参与的固定层级列表，取值见业务常量，空表示不限制"),

		field.JSON("allowed_auto_level", []int64{}).
			Default([]int64{}).
			Comment("允许参与的自动层级列表，取值见业务常量，空表示不限制"),

		field.JSON("allowed_game_ids", []int64{}).
			Default([]int64{}).
			Comment("允许参与的游戏ID列表，空表示不限制"),

		field.JSON("allowed_agent_ids", []int64{}).
			Default([]int64{}).
			Comment("允许参与的代理ID列表，空表示不限制"),

		field.JSON("apply_limit", []int64{}).
			Default([]int64{}).
			Comment("参与活动限制，取值见业务常量，空表示不限制"),

		field.JSON("claim_limit", []int64{}).
			Default([]int64{}).
			Comment("领取奖励限制，可多选，如需完成充值、绑定银行卡等，取值见业务常量，空表示不限制"),

		field.Bool("show_in_list").
			Default(true).
			Comment("是否在活动列表内展示"),

		field.Bool("show_in_sidebar").
			Default(false).
			Comment("是否在侧边栏个人中心入口显示"),

		field.Int64("create_uid").
			Default(0).
			Comment("创建人后台用户 ID"),

		field.Int64("update_uid").
			Default(0).
			Comment("最后更新人后台用户 ID"),

		field.Int("version").
			Default(0).
			NonNegative().
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("配置版本号，乐观锁"),

		field.Int("sort_weight").
			Default(0).
			SchemaType(map[string]string{
				dialect.Postgres: "integer",
			}).
			Comment("展示排序权重，越大越靠前"),
	}
}

// Edges 定义活动表关联关系
func (Activity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("i18ns", ActivityI18n.Type),
	}
}

// Indexes 定义活动表索引
func (Activity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operator_code", "status").
			StorageKey("idx_activity_operator_status"),
	}
}

// Mixin 定义活动表公共字段
func (Activity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
		entmixin.OperatorCodeNotNullMixin{},
	}
}

// Interceptors 定义活动表查询隔离
func (Activity) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			entmixin.FilterOperatorCode(ctx, q)
			return nil
		}),
	}
}

// Annotations 定义活动表数据库注解
func (Activity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("活动配置主表"),
		entsql.Annotation{
			Table: "activity",
		},
	}
}
