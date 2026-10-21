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

// MemberStats 定义会员统计表结构
// 金额按厅结算币种的最小单位存储, last 表示当前结算周期
type MemberStats struct {
	ent.Schema
}

// Fields 定义会员统计表字段
func (MemberStats) Fields() []ent.Field {
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

		// ==================== 累计资金统计 ====================

		field.Int64("total_deposit").
			NonNegative().
			Default(0).
			Comment("累计充值金额"),

		field.Int64("total_deposit_count").
			NonNegative().
			Default(0).
			Comment("累计充值成功次数"),

		field.Int64("total_withdrawal").
			NonNegative().
			Default(0).
			Comment("累计提现金额"),

		field.Int64("total_withdrawal_count").
			NonNegative().
			Default(0).
			Comment("累计提现成功次数"),

		// ==================== 累计游戏统计 ====================

		field.Int64("total_bet").
			NonNegative().
			Default(0).
			Comment("累计投注金额"),

		field.Int64("total_valid_bet").
			NonNegative().
			Default(0).
			Comment("累计有效投注金额"),

		field.Int64("total_water").
			NonNegative().
			Default(0).
			Comment("累计洗码量"),

		field.Int64("total_win_loss").
			Default(0).
			Comment("累计会员净输赢金额, 正数盈利, 负数亏损"),

		// ==================== 当前周期资金统计 ====================

		field.Int64("last_deposit").
			NonNegative().
			Default(0).
			Comment("当前结算周期充值金额"),

		field.Int64("last_deposit_count").
			NonNegative().
			Default(0).
			Comment("当前结算周期充值成功次数"),

		field.Int64("last_withdrawal").
			NonNegative().
			Default(0).
			Comment("当前结算周期提现金额"),

		field.Int64("last_withdrawal_count").
			NonNegative().
			Default(0).
			Comment("当前结算周期提现成功次数"),

		// ==================== 当前周期游戏统计 ====================

		field.Int64("last_bet").
			NonNegative().
			Default(0).
			Comment("当前结算周期投注金额"),

		field.Int64("last_valid_bet").
			NonNegative().
			Default(0).
			Comment("当前结算周期有效投注金额"),

		field.Int64("last_water").
			NonNegative().
			Default(0).
			Comment("当前结算周期洗码量"),

		field.Int64("last_win_loss").
			Default(0).
			Comment("当前结算周期会员净输赢金额, 正数盈利, 负数亏损"),

		// ==================== 结算周期 ====================

		field.Time("period_started_at").
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("当前结算周期开始时间"),
	}
}

// Edges 定义会员统计表关联关系
func (MemberStats) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("stats").
			Field("member_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员统计表索引
func (MemberStats) Indexes() []ent.Index {
	return []ent.Index{

		// 一个会员只有一条实时统计记录
		index.Fields("member_id").
			Unique().
			StorageKey("uk_member_stats_member"),

		// 按厅查询会员统计
		index.Fields("operator_code").
			StorageKey("idx_member_stats_operator_code"),
	}
}

// Mixin 定义会员统计表公共字段
func (MemberStats) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.TimeMixin{},
	}
}

// Interceptors 定义会员统计表查询隔离
func (MemberStats) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks 定义会员统计表写入隔离
func (MemberStats) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
	}
}

// Annotations 定义会员统计表数据库注解
func (MemberStats) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),                // 启用数据库字段注释
		schema.Comment("会员统计表"),                  // 设置数据库表注释
		entsql.Annotation{Table: "member_stats"}, // 设置数据库表名
	}
}
