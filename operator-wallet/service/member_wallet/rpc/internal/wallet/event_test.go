package wallet

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// TestNewWalletEventWithdrawalFreeze 验证完整资金事件的生成
func TestNewWalletEventWithdrawalFreeze(t *testing.T) {
	state := testWalletState()
	state.Balance = 1000
	state.Version = 15

	params := RecordParams{
		OperationID:    uuid.New(),
		BusinessType:   enum.BusinessTypeWithdraw,
		BusinessAction: enum.BusinessActionWithdrawalFreeze,
		BusinessNo:     "W10001",
		Changes: []BalanceChange{
			{BalanceType: enum.BalanceTypeAvailable, ChangeAmount: -300},
			{BalanceType: enum.BalanceTypeFrozen, ChangeAmount: 300},
		},
	}

	prepared, err := prepareRecord(state, params)
	if err != nil {
		t.Fatal(err)
	}

	event := newWalletEvent(state, params, prepared)

	// 校验操作及钱包身份
	if event.SchemaVersion != walletEventSchemaVersion ||
		event.OperationID != params.OperationID ||
		event.WalletID != state.WalletID ||
		event.WalletGlobalID != state.WalletGlobalID ||
		event.MemberGlobalID != state.MemberGlobalID ||
		event.OperatorCode != state.OperatorCode ||
		event.CurrencyCode != state.CurrencyCode {
		t.Fatalf("unexpected event identity: %+v", event)
	}

	// 校验业务来源
	if event.BusinessType != params.BusinessType ||
		event.BusinessAction != params.BusinessAction ||
		event.BusinessNo != params.BusinessNo {
		t.Fatalf("unexpected event business: %+v", event)
	}

	// 一次资金操作只递增一次钱包版本
	if event.PreviousVersion != 15 || event.Version != 16 {
		t.Fatalf("unexpected wallet versions: %d -> %d", event.PreviousVersion, event.Version)
	}

	expectedBefore := walletBalances{
		Balance: 1000,
	}
	expectedAfter := walletBalances{
		Balance:       700,
		FrozenBalance: 300,
	}

	if event.Before != expectedBefore || event.After != expectedAfter {
		t.Fatalf("unexpected balances: before=%+v, after=%+v", event.Before, event.After)
	}

	if len(event.Ledgers) != 2 {
		t.Fatalf("expected 2 ledgers, got %d", len(event.Ledgers))
	}

	// 提现冻结产生两条 Ledger
	expectedLedgers := []walletLedgerEvent{
		{
			LedgerGlobalID: event.Ledgers[0].LedgerGlobalID,
			BalanceType:    enum.BalanceTypeAvailable,
			ChangeAmount:   -300,
			BalanceBefore:  1000,
			BalanceAfter:   700,
		},
		{
			LedgerGlobalID: event.Ledgers[1].LedgerGlobalID,
			BalanceType:    enum.BalanceTypeFrozen,
			ChangeAmount:   300,
			BalanceBefore:  0,
			BalanceAfter:   300,
		},
	}

	if !slices.Equal(event.Ledgers, expectedLedgers) {
		t.Fatalf("unexpected event ledgers: %+v", event.Ledgers)
	}

	// 每条 Ledger 都必须有独立的全局标识
	if event.Ledgers[0].LedgerGlobalID == uuid.Nil ||
		event.Ledgers[1].LedgerGlobalID == uuid.Nil ||
		event.Ledgers[0].LedgerGlobalID == event.Ledgers[1].LedgerGlobalID {
		t.Fatal("invalid ledger global IDs")
	}

	// 事件时间统一使用 UTC, 精度与 PostgreSQL TIMESTAMPTZ(3) 一致
	if event.CreatedAt.Location() != time.UTC || event.CreatedAt.Nanosecond()%int(time.Millisecond) != 0 {
		t.Fatalf("invalid event timestamp: %v", event.CreatedAt)
	}

	// 构造事件不能修改正式钱包状态
	if state.Balance != 1000 || state.FrozenBalance != 0 || state.Version != 15 {
		t.Fatalf("original wallet state was modified: %+v", state)
	}
}

// TestNewWalletEventCopiesLedgers 验证事件不共享候选 Ledger 切片
func TestNewWalletEventCopiesLedgers(t *testing.T) {
	state := testWalletState()
	state.Balance = 500

	params := RecordParams{
		OperationID:    uuid.New(),
		BusinessType:   enum.BusinessTypeDeposit,
		BusinessAction: enum.BusinessActionDepositCredit,
		BusinessNo:     "D10001",
		Changes: []BalanceChange{
			{BalanceType: enum.BalanceTypeAvailable, ChangeAmount: 100},
		},
	}

	prepared, err := prepareRecord(state, params)
	if err != nil {
		t.Fatal(err)
	}

	event := newWalletEvent(state, params, prepared)

	// 修改候选 Ledger, 不应影响已经构造好的事件
	prepared.ledgers[0].ChangeAmount = 200

	if event.Ledgers[0].ChangeAmount != 100 {
		t.Fatalf("event ledger was modified: %+v", event.Ledgers[0])
	}
}

// TestWalletEventOperationResult 验证资金事件可生成对应的操作结果
func TestWalletEventOperationResult(t *testing.T) {
	operationID := uuid.New()
	walletGlobalID := uuid.New()

	event := walletEvent{
		OperationID:    operationID,
		WalletGlobalID: walletGlobalID,
		Version:        16,
		After: walletBalances{
			Balance:          700,
			FrozenBalance:    300,
			UnsettledBalance: 50,
		},
	}

	result := event.operationResult()

	expected := OperationResult{
		OperationID:      operationID,
		WalletGlobalID:   walletGlobalID,
		Version:          16,
		Balance:          700,
		FrozenBalance:    300,
		UnsettledBalance: 50,
	}

	if result != expected {
		t.Fatalf("unexpected operation result: got %+v, want %+v", result, expected)
	}
}
