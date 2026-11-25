package partition

import "testing"

func TestForWallet(t *testing.T) {
	tests := []struct {
		name           string
		walletGlobalID string
		partitionCount int
		want           int
	}{
		{
			name:           "first wallet",
			walletGlobalID: "00000000-0000-0000-0000-000000000001",
			partitionCount: 32,
			want:           18,
		},
		{
			name:           "second wallet",
			walletGlobalID: "00000000-0000-0000-0000-000000000002",
			partitionCount: 32,
			want:           31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ForWallet(tt.walletGlobalID, tt.partitionCount)
			if err != nil {
				t.Fatal(err)
			}

			if got != tt.want {
				t.Fatalf("expected partition %d, got %d", tt.want, got)
			}
		})
	}
}

func TestForWalletInvalid(t *testing.T) {
	tests := []struct {
		name           string
		walletGlobalID string
		partitionCount int
	}{
		{
			name:           "invalid UUID",
			walletGlobalID: "invalid",
			partitionCount: 32,
		},
		{
			name:           "nil UUID",
			walletGlobalID: "00000000-0000-0000-0000-000000000000",
			partitionCount: 32,
		},
		{
			name:           "invalid partition count",
			walletGlobalID: "00000000-0000-0000-0000-000000000001",
			partitionCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ForWallet(tt.walletGlobalID, tt.partitionCount); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
