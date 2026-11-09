// ============================================================================
// 记账器、初始化、记账执行流程
// ============================================================================

package wallet

import (
	"context"
	"errors"
	"fmt"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// recorder 绑定一个事务和一个会员钱包, 负责执行该钱包的资金记账
type recorder struct {
	tx           *ent.Tx           // 当前记账事务
	dao          *dao.Group        // 钱包数据访问
	memberWallet *ent.MemberWallet // 当前会员钱包及余额快照
}

// NewRecorder 创建事务级记账器, memberWallet 必须由调用方在 tx 中查询并锁定
func NewRecorder(tx *ent.Tx, daoGroup *dao.Group, memberWallet *ent.MemberWallet) (*recorder, error) {
	if tx == nil || daoGroup == nil || memberWallet == nil {
		return nil, errors.New("create wallet recorder: missing transaction, DAO or wallet")
	}

	if memberWallet.ID <= 0 {
		return nil, fmt.Errorf("create wallet recorder: invalid wallet ID: %d", memberWallet.ID)
	}

	return &recorder{
		tx:           tx,
		dao:          daoGroup,
		memberWallet: memberWallet,
	}, nil
}

// Record 在当前事务中执行一次资金记账
func (r *recorder) Record(ctx context.Context, in RecordParams) error {
	if r == nil || r.tx == nil || r.dao == nil || r.memberWallet == nil {
		return errors.New("wallet recorder is not initialized")
	}

	if ctx == nil {
		return errors.New("wallet record context is nil")
	}

	// 校验记账请求
	if err := in.validate(); err != nil {
		return err
	}

	// 查询当前操作已有流水, 优先处理幂等
	existing, err := r.dao.MemberWalletLedger.ListByOperationTx(ctx, r.tx, r.memberWallet.ID, in.OperationID)
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

	// 校验当前钱包余额
	if err := validateWalletSnapshot(r.memberWallet); err != nil {
		return err
	}

	// 根据当前钱包余额计算本次变动
	balances := dao.UpdateBalancesParams{
		WalletID:         r.memberWallet.ID,               // 钱包ID
		Balance:          r.memberWallet.Balance,          // 当前可用余额
		FrozenBalance:    r.memberWallet.FrozenBalance,    // 当前冻结余额
		UnsettledBalance: r.memberWallet.UnsettledBalance, // 当前游戏未结算余额
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

		// 计算变动后余额
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
			WalletID:       r.memberWallet.ID,         // 所属钱包ID
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
	if err := r.dao.MemberWallet.UpdateBalancesTx(ctx, r.tx, balances); err != nil {
		return fmt.Errorf("update wallet balances: %w", err)
	}

	// 在同一事务中批量创建资金流水
	if err := r.dao.MemberWalletLedger.CreateManyTx(ctx, r.tx, ledgers); err != nil {
		return fmt.Errorf("create wallet ledgers: %w", err)
	}

	// 同步当前事务中的钱包快照
	r.memberWallet.Balance = balances.Balance
	r.memberWallet.FrozenBalance = balances.FrozenBalance
	r.memberWallet.UnsettledBalance = balances.UnsettledBalance

	return nil
}

// validateWalletSnapshot 校验当前钱包余额
func validateWalletSnapshot(memberWallet *ent.MemberWallet) error {
	if memberWallet.Balance < 0 || memberWallet.FrozenBalance < 0 || memberWallet.UnsettledBalance < 0 {
		return fmt.Errorf("wallet contains negative balance: wallet_id=%d", memberWallet.ID)
	}

	return nil
}

// matchExistingRecords 判断历史流水是否与本次操作完全一致
func matchExistingRecords(existing []*ent.MemberWalletLedger, in RecordParams) bool {
	// 历史流水数量必须与本次余额变动数量一致
	if len(existing) != len(in.Changes) {
		return false
	}

	// 按余额类型建立变动金额映射, 避免依赖流水查询顺序
	changes := make(map[enum.BalanceType]int64, len(in.Changes))
	for _, change := range in.Changes {
		changes[change.BalanceType] = change.ChangeAmount
	}

	for _, ledger := range existing {
		// 校验流水所属操作及业务来源是否一致
		if ledger == nil ||
			ledger.OperationID != in.OperationID ||
			ledger.BusinessType != int64(in.BusinessType) ||
			ledger.BusinessAction != int64(in.BusinessAction) ||
			ledger.BusinessNo != in.BusinessNo {
			return false
		}

		// 校验余额类型及变动金额
		balanceType := enum.BalanceType(ledger.BalanceType)
		amount, exists := changes[balanceType]
		if !exists || amount != ledger.ChangeAmount {
			return false
		}

		// 移除已匹配的变动, 防止重复匹配
		delete(changes, balanceType)
	}

	// 所有余额变动均匹配, 才视为同一操作
	return len(changes) == 0
}
