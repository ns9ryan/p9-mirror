package schema

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"oa.98ent.com/p9/common/entmixin"
	"oa.98ent.com/p9/operator-wallet/ent/intercept"
)

// MemberWalletLedger 定义会员钱包资金流水表结构
type MemberWalletLedger struct {
	ent.Schema
}

// Fields 定义会员钱包资金流水表字段
func (MemberWalletLedger) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 流水标识 ====================

		field.UUID("global_id", uuid.UUID{}).
			Unique().
			Immutable().
			Default(uuid.New).
			Comment("资金流水全局唯一标识"),

		// ==================== 钱包归属 ====================

		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("wallet_id").
			Positive().
			Immutable().
			Comment("所属钱包ID"),

		// ==================== 余额变动 ====================

		field.Int64("balance_type").
			Positive().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Immutable().
			Comment("余额类型: 1可用余额, 2冻结余额, 3未结算余额"),

		field.Int64("change_amount").
			Immutable().
			Comment("变动金额, 正数增加, 负数减少"),

		field.Int64("balance_before").
			NonNegative().
			Immutable().
			Comment("变动前余额"),

		field.Int64("balance_after").
			NonNegative().
			Immutable().
			Comment("变动后余额"),

		// ==================== 操作标识 ====================

		field.UUID("operation_id", uuid.UUID{}).
			Immutable().
			Comment("钱包资金操作唯一标识"),

		// ==================== 业务来源 ====================

		field.Int64("business_type").
			Positive().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Immutable().
			Comment("一级业务类型: 1存款, 2提现, 3游戏, 4活动, 5任务, 6佣金, 7系统奖励"),

		field.Int64("business_action").
			Positive().
			SchemaType(map[string]string{
				dialect.Postgres: "smallint",
			}).
			Immutable().
			Comment("具体资金操作类型"),

		field.String("business_no").
			NotEmpty().
			MaxLen(128).
			Immutable().
			Comment("原始业务单号"),

		// ==================== 时间 ====================

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz(3)",
			}).
			Comment("流水创建时间"),
	}
}

// Edges 定义会员钱包资金流水表关联关系
func (MemberWalletLedger) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("wallet", MemberWallet.Type).
			Ref("ledgers").
			Field("wallet_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes 定义会员钱包资金流水表索引
func (MemberWalletLedger) Indexes() []ent.Index {
	return []ent.Index{

		// 同一次操作对同一种余额只能产生一条流水
		index.Fields("operator_code", "wallet_id", "operation_id", "balance_type").
			Unique().
			StorageKey("uk_member_wallet_ledger_operation"),

		// 同一厅的一笔存款订单只能入账一次本金
		index.Fields("operator_code", "business_no").
			Unique().
			StorageKey("uk_member_wallet_ledger_deposit_credit").
			Annotations(
				entsql.IndexWhere("business_type = 1 AND business_action = 1001"),
			),

		// 同一厅的一笔游戏下注只能记账一次
		index.Fields("operator_code", "business_no", "balance_type").
			Unique().
			StorageKey("uk_member_wallet_ledger_game_bet").
			Annotations(
				entsql.IndexWhere("business_type = 3 AND business_action = 3001"),
			),

		// 查询钱包资金明细, 支持按时间倒序分页
		index.Fields("operator_code", "wallet_id", "created_at", "id").
			StorageKey("idx_member_wallet_ledger_wallet_created"),

		// 根据原始业务单号查询资金流水
		index.Fields("operator_code", "business_type", "business_no").
			StorageKey("idx_member_wallet_ledger_business_no"),
	}
}

// Mixin 定义会员钱包资金流水表公共字段
func (MemberWalletLedger) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
	}
}

// Interceptors 定义会员钱包资金流水表查询隔离
func (MemberWalletLedger) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		entmixin.RequiredOperatorCodeInterceptor(intercept.NewQuery),
	}
}

// Hooks 定义会员钱包资金流水表写入隔离
func (MemberWalletLedger) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
		rejectLedgerMutationHook(),
	}
}

// Annotations 定义会员钱包资金流水表数据库注解
func (MemberWalletLedger) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("会员钱包资金流水表"),
		entsql.Annotation{
			Table: "member_wallet_ledger",
			Checks: map[string]string{
				"ck_member_wallet_ledger_balance_type":        "balance_type > 0",                               // 余额类型必须为正数
				"ck_member_wallet_ledger_business_type":       "business_type > 0",                              // 一级业务类型必须为正数
				"ck_member_wallet_ledger_business_action":     "business_action > 0",                            // 具体操作类型必须为正数
				"ck_member_wallet_ledger_change_amount":       "change_amount <> 0",                             // 变动金额不能为零
				"ck_member_wallet_ledger_balance_before":      "balance_before >= 0",                            // 变动前余额不能为负数
				"ck_member_wallet_ledger_balance_after":       "balance_after >= 0",                             // 变动后余额不能为负数
				"ck_member_wallet_ledger_balance_consistency": "balance_after = balance_before + change_amount", // 保证余额计算一致
			},
		},
	}
}

// rejectLedgerMutationHook 禁止修改和删除已生成的资金流水
func rejectLedgerMutationHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op() != ent.OpCreate {
				return nil, errors.New("member wallet ledger is append-only")
			}
			return next.Mutate(ctx, m)
		})
	}
}
