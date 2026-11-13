package wallet

import (
	"errors"
	"math"
	"slices"
	"testing"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// TestPrepareRecordWithdrawalFreeze 验证提现冻结的余额计算和流水生成
func TestPrepareRecordWithdrawalFreeze(t *testing.T) {
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

	result, err := prepareRecord(state, params)
	if err != nil {
		t.Fatal(err)
	}

	// 一次资金操作产生两条流水, 但钱包版本只增加一次
	expectedState := state
	expectedState.Balance = 700
	expectedState.FrozenBalance = 300
	expectedState.Version = 16

	if result.nextState != expectedState {
		t.Fatalf("unexpected next state: got %+v, want %+v", result.nextState, expectedState)
	}

	expectedLedgers := []preparedLedger{
		{
			BalanceType:   enum.BalanceTypeAvailable,
			ChangeAmount:  -300,
			BalanceBefore: 1000,
			BalanceAfter:  700,
		},
		{
			BalanceType:   enum.BalanceTypeFrozen,
			ChangeAmount:  300,
			BalanceBefore: 0,
			BalanceAfter:  300,
		},
	}

	if !slices.Equal(result.ledgers, expectedLedgers) {
		t.Fatalf("unexpected ledgers: got %+v, want %+v", result.ledgers, expectedLedgers)
	}

	// 计算成功也不能提前修改 Actor 的正式余额和版本
	if state.Balance != 1000 || state.FrozenBalance != 0 || state.Version != 15 {
		t.Fatalf("original wallet state was modified: %+v", state)
	}
}

// TestPrepareRecordGameSettlement 验证游戏部分结算的资金计算
func TestPrepareRecordGameSettlement(t *testing.T) {
	state := testWalletState()
	state.Balance = 200
	state.UnsettledBalance = 100
	state.Version = 4

	params := RecordParams{
		OperationID:    uuid.New(),
		BusinessType:   enum.BusinessTypeGame,
		BusinessAction: enum.BusinessActionGameSettlement,
		BusinessNo:     "G10001",
		Changes: []BalanceChange{
			{BalanceType: enum.BalanceTypeUnsettled, ChangeAmount: -40},
			{BalanceType: enum.BalanceTypeAvailable, ChangeAmount: 60},
		},
	}

	result, err := prepareRecord(state, params)
	if err != nil {
		t.Fatal(err)
	}

	if result.nextState.Balance != 260 ||
		result.nextState.UnsettledBalance != 60 ||
		result.nextState.Version != 5 {
		t.Fatalf("unexpected settlement result: %+v", result.nextState)
	}

	if len(result.ledgers) != 2 {
		t.Fatalf("expected 2 ledgers, got %d", len(result.ledgers))
	}

	// 原始未结算金额不应被候选计算修改
	if state.UnsettledBalance != 100 || state.Version != 4 {
		t.Fatalf("original wallet state was modified: %+v", state)
	}
}

// TestPrepareRecordRejectsInvalidChanges 验证非法操作不会产生候选结果
func TestPrepareRecordRejectsInvalidChanges(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*State, *RecordParams)
		wantErr error
	}{
		{
			name: "insufficient balance",
			modify: func(_ *State, p *RecordParams) {
				p.Changes[0].ChangeAmount = -101
			},
			wantErr: ErrInsufficientBalance,
		},
		{
			name: "balance overflow",
			modify: func(s *State, p *RecordParams) {
				s.Balance = math.MaxInt64
				p.Changes[0].ChangeAmount = 1
			},
			wantErr: ErrBalanceOverflow,
		},
		{
			name: "version overflow",
			modify: func(s *State, _ *RecordParams) {
				s.Version = math.MaxInt64
			},
			wantErr: ErrWalletVersionOverflow,
		},
		{
			name: "invalid wallet state",
			modify: func(s *State, _ *RecordParams) {
				s.FrozenBalance = -1
			},
			wantErr: ErrInvalidWalletState,
		},
		{
			name: "empty operation ID",
			modify: func(_ *State, p *RecordParams) {
				p.OperationID = uuid.Nil
			},
			wantErr: ErrInvalidRecordParams,
		},
		{
			name: "duplicate balance type",
			modify: func(_ *State, p *RecordParams) {
				p.Changes = append(p.Changes, BalanceChange{
					BalanceType:  enum.BalanceTypeAvailable,
					ChangeAmount: 10,
				})
			},
			wantErr: ErrInvalidRecordParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := testWalletState()
			state.Balance = 100

			params := RecordParams{
				OperationID:    uuid.New(),
				BusinessType:   enum.BusinessTypeDeposit,
				BusinessAction: enum.BusinessActionDepositCredit,
				BusinessNo:     "D10001",
				Changes: []BalanceChange{
					{BalanceType: enum.BalanceTypeAvailable, ChangeAmount: -10},
				},
			}

			tt.modify(&state, &params)
			original := state

			result, err := prepareRecord(state, params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}

			// 校验失败不能返回部分计算结果
			if result.nextState != (State{}) || len(result.ledgers) != 0 {
				t.Fatalf("unexpected prepared result: %+v", result)
			}

			if state != original {
				t.Fatalf("original wallet state was modified: %+v", state)
			}
		})
	}
}
