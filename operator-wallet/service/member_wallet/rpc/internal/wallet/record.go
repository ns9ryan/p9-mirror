package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// 统一记账业务错误
var (
	ErrInvalidRecordParams = errors.New("invalid wallet record parameters") // 记账参数不合法
	ErrInsufficientBalance = errors.New("insufficient wallet balance")      // 钱包余额不足
	ErrBalanceOverflow     = errors.New("wallet balance overflow")          // 钱包余额溢出
	ErrOperationConflict   = errors.New("wallet operation conflict")        // 钱包操作标识冲突
)

// BalanceChange 单种钱包余额变动
type BalanceChange struct {
	BalanceType  enum.BalanceType // 余额类型
	ChangeAmount int64            // 变动金额, 正数增加, 负数减少
}

// RecordParams 统一资金记账参数
type RecordParams struct {
	// 操作与业务来源
	OperationID    uuid.UUID           // 钱包操作唯一标识
	BusinessType   enum.BusinessType   // 一级业务类型
	BusinessAction enum.BusinessAction // 具体资金操作类型
	BusinessNo     string              // 原始业务单号

	Changes []BalanceChange // 本次操作的余额变动集合
}

// RecordTx 在当前事务中执行统一资金记账
// 调用方负责事务管理, 并传入同一事务中已锁定的钱包
func RecordTx(ctx context.Context, tx *ent.Tx, d *dao.Group, memberWallet *ent.MemberWallet, in RecordParams) error {
	// 校验调用参数和余额变动集合
	if ctx == nil || tx == nil || d == nil || memberWallet == nil {
		return invalidRecordParams("missing transaction, DAO or wallet")
	}
	if memberWallet.ID <= 0 {
		return invalidRecordParams("invalid wallet ID")
	}
	if err := validateRecordParams(memberWallet, in); err != nil {
		return err
	}

	// 查询当前操作已有流水, 优先处理幂等
	existing, err := d.MemberWalletLedger.ListByOperationTx(ctx, tx, memberWallet.ID, in.OperationID)
	if err != nil {
		return fmt.Errorf("query wallet operation ledgers: %w", err)
	}

	// 已存在的操作必须整组一致, 否则视为冲突
	if len(existing) > 0 {
		if !matchExistingRecords(existing, in) {
			return ErrOperationConflict
		}
		return nil
	}

	// 根据当前钱包余额计算本次变动
	balances := dao.UpdateBalancesParams{
		WalletID:         memberWallet.ID,               // 钱包ID
		Balance:          memberWallet.Balance,          // 当前可用余额
		FrozenBalance:    memberWallet.FrozenBalance,    // 当前冻结余额
		UnsettledBalance: memberWallet.UnsettledBalance, // 当前游戏未结算余额
	}

	ledgers := make([]dao.CreateLedgerParams, 0, len(in.Changes))

	for _, change := range in.Changes {
		var before int64

		// 获取对应类型的变动前余额
		switch change.BalanceType {
		case enum.BalanceTypeAvailable:
			before = balances.Balance
		case enum.BalanceTypeFrozen:
			before = balances.FrozenBalance
		case enum.BalanceTypeUnsettled:
			before = balances.UnsettledBalance
		}

		// 计算变动后余额, 同时检查透支和溢出
		after, err := applyBalanceChange(before, change.ChangeAmount, change.BalanceType)
		if err != nil {
			return err
		}

		// 更新对应类型的最终余额
		switch change.BalanceType {
		case enum.BalanceTypeAvailable:
			balances.Balance = after
		case enum.BalanceTypeFrozen:
			balances.FrozenBalance = after
		case enum.BalanceTypeUnsettled:
			balances.UnsettledBalance = after
		}

		// 记录本次余额变动
		ledgers = append(ledgers, dao.CreateLedgerParams{
			WalletID:       memberWallet.ID,           // 所属钱包ID
			BalanceType:    int64(change.BalanceType), // 余额类型
			ChangeAmount:   change.ChangeAmount,       // 变动金额
			BalanceBefore:  before,                    // 变动前余额
			BalanceAfter:   after,                     // 变动后余额
			OperationID:    in.OperationID,            // 钱包操作唯一标识
			BusinessType:   int64(in.BusinessType),    // 一级业务类型
			BusinessAction: int64(in.BusinessAction),  // 具体资金操作类型
			BusinessNo:     in.BusinessNo,             // 原始业务单号
		})
	}

	// 一次更新钱包的三种余额
	if err := d.MemberWallet.UpdateBalancesTx(ctx, tx, balances); err != nil {
		return fmt.Errorf("update wallet balances: %w", err)
	}

	// 在同一事务中批量创建资金流水
	if err := d.MemberWalletLedger.CreateManyTx(ctx, tx, ledgers); err != nil {
		return fmt.Errorf("create wallet ledgers: %w", err)
	}

	// 同步当前事务中的钱包快照
	memberWallet.Balance = balances.Balance
	memberWallet.FrozenBalance = balances.FrozenBalance
	memberWallet.UnsettledBalance = balances.UnsettledBalance

	return nil
}

// invalidRecordParams 统一构造记账参数错误
func invalidRecordParams(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRecordParams, reason)
}

// validateRecordParams 校验统一记账参数
func validateRecordParams(memberWallet *ent.MemberWallet, in RecordParams) error {
	// 钱包余额必须符合数据库约束
	if memberWallet.Balance < 0 || memberWallet.FrozenBalance < 0 || memberWallet.UnsettledBalance < 0 {
		return invalidRecordParams("wallet balance cannot be negative")
	}

	// 校验操作标识
	if in.OperationID == uuid.Nil {
		return invalidRecordParams("empty operation ID")
	}

	// 校验业务主类型
	switch in.BusinessType {
	case enum.BusinessTypeDeposit,
		enum.BusinessTypeWithdraw,
		enum.BusinessTypeGame,
		enum.BusinessTypeActivity,
		enum.BusinessTypeTask,
		enum.BusinessTypeCommission,
		enum.BusinessTypeSystemReward:
	default:
		return invalidRecordParams("unsupported business type")
	}

	// BusinessAction 按 BusinessType 分段编码
	if in.BusinessAction <= 0 ||
		int64(in.BusinessAction)/1000 != int64(in.BusinessType) {
		return invalidRecordParams("business action does not match business type")
	}

	if in.BusinessNo == "" || strings.TrimSpace(in.BusinessNo) != in.BusinessNo || utf8.RuneCountInString(in.BusinessNo) > 128 {
		return invalidRecordParams("invalid business number")
	}

	// 一次操作最多涉及三种余额, 同一种余额只能出现一次
	if len(in.Changes) == 0 || len(in.Changes) > 3 {
		return invalidRecordParams("invalid balance changes")
	}

	seen := make(map[enum.BalanceType]struct{}, len(in.Changes))

	for _, change := range in.Changes {
		switch change.BalanceType {
		case enum.BalanceTypeAvailable, enum.BalanceTypeFrozen, enum.BalanceTypeUnsettled:
		default:
			return invalidRecordParams("unsupported balance type")
		}

		if change.ChangeAmount == 0 {
			return invalidRecordParams("zero balance change")
		}

		if _, exists := seen[change.BalanceType]; exists {
			return invalidRecordParams("duplicate balance type")
		}
		seen[change.BalanceType] = struct{}{}
	}

	return nil
}

// matchExistingRecords 判断历史流水是否与本次操作完全一致
func matchExistingRecords(existing []*ent.MemberWalletLedger, in RecordParams) bool {
	if len(existing) != len(in.Changes) {
		return false
	}

	// 按余额类型匹配, 不依赖切片顺序
	changes := make(map[enum.BalanceType]int64, len(in.Changes))
	for _, change := range in.Changes {
		changes[change.BalanceType] = change.ChangeAmount
	}

	for _, ledger := range existing {
		if ledger == nil ||
			ledger.OperationID != in.OperationID ||
			ledger.BusinessType != int64(in.BusinessType) ||
			ledger.BusinessAction != int64(in.BusinessAction) ||
			ledger.BusinessNo != in.BusinessNo {
			return false
		}

		balanceType := enum.BalanceType(ledger.BalanceType)
		amount, exists := changes[balanceType]
		if !exists || amount != ledger.ChangeAmount {
			return false
		}

		delete(changes, balanceType)
	}

	return len(changes) == 0
}

// applyBalanceChange 计算余额变动并检查透支和溢出
func applyBalanceChange(balance, change int64, balanceType enum.BalanceType) (int64, error) {
	if balance < 0 {
		return 0, invalidRecordParams("wallet balance cannot be negative")
	}
	if change == 0 {
		return 0, invalidRecordParams("zero balance change")
	}

	// 增加余额时检查 int64 上限
	if change > 0 && balance > math.MaxInt64-change {
		return 0, fmt.Errorf("%w: balance_type=%d, balance=%d, change=%d", ErrBalanceOverflow, balanceType, balance, change)
	}

	// 负数直接比较, 避免对 math.MinInt64 取反造成溢出
	if change < 0 && change < -balance {
		return 0, fmt.Errorf("%w: balance_type=%d, balance=%d, change=%d", ErrInsufficientBalance, balanceType, balance, change)
	}

	return balance + change, nil
}
