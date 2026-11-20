package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-crm/ent/intercept"
	"oa.98ent.com/p9/operator-crm/ent/schema/mixins"
)

// MemberPeriodStats 定义会员结算周期历史统计表结构
// 每个会员每个周期一条记录, 金额按厅结算币种的最小单位存储
type MemberPeriodStats struct {
	ent.Schema
}

// Fields 定义会员结算周期历史统计表字段
func (MemberPeriodStats) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 会员关联 ====================

		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("所属会员ID"),

		// ==================== 结算周期 ====================

		field.Time("period_started_at").
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("结算周期开始时间, 包含"),

		field.Time("period_ended_at").
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("结算周期结束时间, 不包含"),

		// ==================== 资金统计 ====================

		field.Int64("deposit").
			NonNegative().
			Default(0).
			Comment("本周期充值金额"),

		field.Int64("deposit_count").
			NonNegative().
			Default(0).
			Comment("本周期充值成功次数"),

		field.Int64("withdrawal").
			NonNegative().
			Default(0).
			Comment("本周期提现金额"),

		field.Int64("withdrawal_count").
			NonNegative().
			Default(0).
			Comment("本周期提现成功次数"),

		// ==================== 游戏统计 ====================

		field.Int64("bet").
			NonNegative().
			Default(0).
			Comment("本周期投注金额"),

		field.Int64("valid_bet").
			NonNegative().
			Default(0).
			Comment("本周期有效投注金额"),

		field.Int64("water").
			NonNegative().
			Default(0).
			Comment("本周期洗码量"),

		field.Int64("win_loss").
			Default(0).
			Comment("本周期会员净输赢金额, 正数盈利, 负数亏损"),
	}
}

// Edges 定义会员结算周期历史统计表关联关系
func (MemberPeriodStats) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("period_stats").
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员结算周期历史统计表索引
func (MemberPeriodStats) Indexes() []ent.Index {
	return []ent.Index{

		// 同一会员的同一结算周期只能归档一次
		index.Fields("member_id", "period_started_at").
			Unique().
			StorageKey("uk_member_period_stats_member_period"),

		// 按厅和结算周期查询历史统计
		index.Fields("operator_code", "period_started_at").
			StorageKey("idx_member_period_stats_operator_period"),
	}
}

// Mixin 定义会员结算周期历史统计表公共字段
func (MemberPeriodStats) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors 定义会员结算周期历史统计表查询隔离
func (MemberPeriodStats) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks 定义会员结算周期历史统计表写入隔离
func (MemberPeriodStats) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
	}
}

// Annotations 定义会员结算周期历史统计表数据库注解
func (MemberPeriodStats) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                       // 启用数据库字段注释
		schema.Comment("会员结算周期历史统计表"),                   // 设置数据库表注释
		entsql.Annotation{Table: "member_period_stats"}, // 设置数据库表名
	}
}
