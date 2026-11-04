package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"oa.98ent.com/p9/common/entmixin"
)

// MemberWallet 定义会员钱包表结构
type MemberWallet struct {
	ent.Schema
}

// Fields 定义会员钱包表字段
func (MemberWallet) Fields() []ent.Field {
	return []ent.Field{

		// ==================== 钱包标识与会员归属 ====================

		field.UUID("global_id", uuid.UUID{}).
			Unique().
			Immutable().
			Default(uuid.New).
			Comment("钱包全局唯一标识"),

		field.String("operator_code").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("所属厅业务编码"),

		field.Int64("member_id").
			Positive().
			Immutable().
			Comment("CRM会员ID"),

		field.UUID("member_global_id", uuid.UUID{}).
			Unique().
			Immutable().
			Comment("CRM会员全局唯一标识"),

		// ==================== 钱包币种 ====================

		field.String("currency_code").
			NotEmpty().
			MaxLen(16).
			Immutable().
			Comment("钱包结算币种编码"),

		// ==================== 账户余额, 按币种最小单位存储 ====================

		field.Int64("balance").
			NonNegative().
			Default(0).
			Comment("可用余额"),

		field.Int64("frozen_balance").
			NonNegative().
			Default(0).
			Comment("冻结余额"),

		field.Int64("unsettled_balance").
			NonNegative().
			Default(0).
			Comment("游戏未结算余额"),
	}
}

// Indexes 定义会员钱包表索引
func (MemberWallet) Indexes() []ent.Index {
	return []ent.Index{
		// 同一个厅的会员只能拥有一个钱包
		index.Fields("operator_code", "member_id").
			Unique().
			StorageKey("uk_member_wallet_operator_member"),
	}
}

// Mixin 定义会员钱包表公共字段
func (MemberWallet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entmixin.IDMixin{},
		entmixin.TimeMixin{},
	}
}

// Hooks 定义会员钱包表写入隔离
func (MemberWallet) Hooks() []ent.Hook {
	return []ent.Hook{
		entmixin.RequiredOperatorCodeHook(),
	}
}

// Annotations 定义会员钱包表数据库注解
func (MemberWallet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("会员钱包表"),
		entsql.Annotation{
			Table: "member_wallet",
			Checks: map[string]string{
				"ck_member_wallet_balance":           "balance >= 0",           // 可用余额不能为负数
				"ck_member_wallet_frozen_balance":    "frozen_balance >= 0",    // 冻结余额不能为负数
				"ck_member_wallet_unsettled_balance": "unsettled_balance >= 0", // 游戏未结算余额不能为负数
			},
		},
	}
}
