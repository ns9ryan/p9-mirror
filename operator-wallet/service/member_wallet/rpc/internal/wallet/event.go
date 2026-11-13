package wallet

import (
	"time"

	"github.com/google/uuid"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

const walletEventSchemaVersion = 1

// walletBalances 表示某个钱包版本的三类余额快照
type walletBalances struct {
	Balance          int64 // 可用余额
	FrozenBalance    int64 // 冻结余额
	UnsettledBalance int64 // 游戏未结算余额
}

// walletLedgerEvent 表示资金事件中的单条流水
type walletLedgerEvent struct {
	LedgerGlobalID uuid.UUID        // 流水全局唯一标识
	BalanceType    enum.BalanceType // 余额类型
	ChangeAmount   int64            // 变动金额
	BalanceBefore  int64            // 变动前余额
	BalanceAfter   int64            // 变动后余额
}

// walletEvent 表示一次完整的资金操作事件
// 一次操作只生成一条事件, 事件内可包含多条 Ledger
type walletEvent struct {
	SchemaVersion int       // 事件结构版本
	OperationID   uuid.UUID // 本次资金操作唯一标识

	// 钱包与会员归属
	WalletID       int64     // 数据库钱包ID
	WalletGlobalID uuid.UUID // 钱包全局唯一标识
	MemberGlobalID uuid.UUID // 会员全局唯一标识
	OperatorCode   string    // 所属厅业务编码
	CurrencyCode   string    // 钱包结算币种编码

	// 业务来源
	BusinessType   enum.BusinessType   // 一级业务类型
	BusinessAction enum.BusinessAction // 具体资金操作类型
	BusinessNo     string              // 原始业务订单号

	// 钱包版本与余额
	PreviousVersion int64          // 操作前钱包版本
	Version         int64          // 操作后钱包版本
	Before          walletBalances // 操作前余额
	After           walletBalances // 操作后余额

	Ledgers   []walletLedgerEvent // 本次操作产生的全部资金流水
	CreatedAt time.Time           // 资金事件生成时间
}

// newWalletEvent 根据已校验的候选结果构造完整资金事件
// 必须在 prepareRecord 成功后调用, 不修改当前钱包状态
func newWalletEvent(state State, params RecordParams, prepared preparedRecord) walletEvent {
	ledgers := make([]walletLedgerEvent, 0, len(prepared.ledgers))

	for _, ledger := range prepared.ledgers {
		ledgers = append(ledgers, walletLedgerEvent{
			LedgerGlobalID: uuid.New(),           // 流水全局唯一标识
			BalanceType:    ledger.BalanceType,   // 余额类型
			ChangeAmount:   ledger.ChangeAmount,  // 变动金额
			BalanceBefore:  ledger.BalanceBefore, // 变动前余额
			BalanceAfter:   ledger.BalanceAfter,  // 变动后余额
		})
	}

	return walletEvent{
		SchemaVersion: walletEventSchemaVersion, // 事件结构版本
		OperationID:   params.OperationID,       // 本次资金操作唯一标识

		// 钱包与会员归属
		WalletID:       state.WalletID,       // 数据库钱包ID
		WalletGlobalID: state.WalletGlobalID, // 钱包全局唯一标识
		MemberGlobalID: state.MemberGlobalID, // 会员全局唯一标识
		OperatorCode:   state.OperatorCode,   // 所属厅业务编码
		CurrencyCode:   state.CurrencyCode,   // 钱包结算币种编码

		// 业务来源
		BusinessType:   params.BusinessType,   // 一级业务类型
		BusinessAction: params.BusinessAction, // 具体资金操作类型
		BusinessNo:     params.BusinessNo,     // 原始业务订单号

		// 钱包版本与余额
		PreviousVersion: state.Version,                         // 操作前钱包版本
		Version:         prepared.nextState.Version,            // 操作后钱包版本
		Before:          balancesFromState(state),              // 操作前余额
		After:           balancesFromState(prepared.nextState), // 操作后余额

		Ledgers:   ledgers,                                     // 本次操作产生的全部资金流水
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond), // 资金事件生成时间
	}
}

// balancesFromState 提取钱包的三类余额
func balancesFromState(state State) walletBalances {
	return walletBalances{
		Balance:          state.Balance,          // 可用余额
		FrozenBalance:    state.FrozenBalance,    // 冻结余额
		UnsettledBalance: state.UnsettledBalance, // 游戏未结算余额
	}
}

// operationResult 返回本次资金操作完成时的钱包状态
func (e walletEvent) operationResult() OperationResult {
	return OperationResult{
		// 操作与钱包标识
		OperationID:    e.OperationID,    // 资金操作唯一标识
		WalletGlobalID: e.WalletGlobalID, // 钱包全局唯一标识

		// 操作完成时的钱包状态
		Version:          e.Version,                // 本次操作完成后的钱包版本
		Balance:          e.After.Balance,          // 可用余额
		FrozenBalance:    e.After.FrozenBalance,    // 冻结余额
		UnsettledBalance: e.After.UnsettledBalance, // 游戏未结算余额
	}
}
