// ============================================================================
// 资金流水匹配
// ============================================================================

package memberwalletlogic

import (
	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// matchBusinessLedgerChanges 校验业务流水与预期余额变动是否完全一致
func matchBusinessLedgerChanges(
	existing []*ent.MemberWalletLedger,
	walletID int64,
	businessType enum.BusinessType,
	businessAction enum.BusinessAction,
	businessNo string,
	expectedChanges map[enum.BalanceType]int64,
) bool {
	// 流水数量必须与预期余额变动数量一致
	if len(expectedChanges) == 0 || len(existing) != len(expectedChanges) || existing[0] == nil {
		return false
	}

	// 所有流水必须属于同一次钱包操作
	operationID := existing[0].OperationID
	if operationID == uuid.Nil {
		return false
	}

	// 复制预期变动, 避免修改调用方传入的 map
	changes := make(map[enum.BalanceType]int64, len(expectedChanges))
	for balanceType, amount := range expectedChanges {
		if !balanceType.IsValid() || amount == 0 {
			return false
		}
		changes[balanceType] = amount
	}

	for _, ledger := range existing {
		if ledger == nil ||
			ledger.WalletID != walletID ||
			ledger.OperationID != operationID ||
			ledger.BusinessType != int64(businessType) ||
			ledger.BusinessAction != int64(businessAction) ||
			ledger.BusinessNo != businessNo {
			return false
		}

		// 每种余额只能匹配一条流水
		balanceType := enum.BalanceType(ledger.BalanceType)
		amount, exists := changes[balanceType]
		if !exists || amount != ledger.ChangeAmount {
			return false
		}

		delete(changes, balanceType)
	}

	return len(changes) == 0
}
