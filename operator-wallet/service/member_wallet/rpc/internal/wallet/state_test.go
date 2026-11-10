package wallet

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestWalletStateValidate(t *testing.T) {
	base := State{
		WalletID:       1,
		WalletGlobalID: uuid.New(),
		MemberGlobalID: uuid.New(),
		OperatorCode:   "op001",
		CurrencyCode:   "USDT",
	}

	tests := []struct {
		name    string
		mutate  func(*State)
		wantErr bool
	}{
		{
			name: "valid state",
		},
		{
			name: "invalid wallet ID",
			mutate: func(s *State) {
				s.WalletID = 0
			},
			wantErr: true,
		},
		{
			name: "empty wallet global ID",
			mutate: func(s *State) {
				s.WalletGlobalID = uuid.Nil
			},
			wantErr: true,
		},
		{
			name: "empty member global ID",
			mutate: func(s *State) {
				s.MemberGlobalID = uuid.Nil
			},
			wantErr: true,
		},
		{
			name: "invalid operator code",
			mutate: func(s *State) {
				s.OperatorCode = " op001"
			},
			wantErr: true,
		},
		{
			name: "operator code too long",
			mutate: func(s *State) {
				s.OperatorCode = strings.Repeat("a", 65)
			},
			wantErr: true,
		},
		{
			name: "invalid currency code",
			mutate: func(s *State) {
				s.CurrencyCode = ""
			},
			wantErr: true,
		},
		{
			name: "currency code too long",
			mutate: func(s *State) {
				s.CurrencyCode = strings.Repeat("a", 17)
			},
			wantErr: true,
		},
		{
			name: "negative available balance",
			mutate: func(s *State) {
				s.Balance = -1
			},
			wantErr: true,
		},
		{
			name: "negative frozen balance",
			mutate: func(s *State) {
				s.FrozenBalance = -1
			},
			wantErr: true,
		},
		{
			name: "negative unsettled balance",
			mutate: func(s *State) {
				s.UnsettledBalance = -1
			},
			wantErr: true,
		},
		{
			name: "negative version",
			mutate: func(s *State) {
				s.Version = -1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := base

			if tt.mutate != nil {
				tt.mutate(&state)
			}

			err := state.Validate()

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidWalletState) {
					t.Fatalf("expected ErrInvalidWalletState, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
