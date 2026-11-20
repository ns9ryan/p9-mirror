package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// TestKafkaProducerBuildRecord 验证资金事件的Partition、Key和JSON内容
func TestKafkaProducerBuildRecord(t *testing.T) {
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
	producer := &kafkaProducer{topic: "wallet-test"}

	record, err := producer.buildRecord(3, event)
	if err != nil {
		t.Fatal(err)
	}

	if record.Topic != "wallet-test" || record.Partition != 3 {
		t.Fatalf("unexpected Kafka destination: topic=%s, partition=%d", record.Topic, record.Partition)
	}
	if !bytes.Equal(record.Key, state.WalletGlobalID[:]) {
		t.Fatal("Kafka key does not match wallet global ID")
	}

	// 解析JSON, 验证固定字段名和完整的资金结果
	var decoded struct {
		SchemaVersion   int       `json:"schema_version"`
		OperationID     uuid.UUID `json:"operation_id"`
		WalletGlobalID  uuid.UUID `json:"wallet_global_id"`
		PreviousVersion int64     `json:"previous_version"`
		Version         int64     `json:"version"`

		After struct {
			Balance       int64 `json:"balance"`
			FrozenBalance int64 `json:"frozen_balance"`
		} `json:"after"`

		Ledgers []walletLedgerEvent `json:"ledgers"`
	}

	if err := json.Unmarshal(record.Value, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.SchemaVersion != walletEventSchemaVersion ||
		decoded.OperationID != params.OperationID ||
		decoded.WalletGlobalID != state.WalletGlobalID ||
		decoded.PreviousVersion != 15 ||
		decoded.Version != 16 ||
		decoded.After.Balance != 700 ||
		decoded.After.FrozenBalance != 300 ||
		len(decoded.Ledgers) != 2 {
		t.Fatalf("unexpected Kafka event: %+v", decoded)
	}

	// 确认JSON使用固定字段名, 而不是Go结构体字段名
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(record.Value, &fields); err != nil {
		t.Fatal(err)
	}
	if _, exists := fields["SchemaVersion"]; exists {
		t.Fatal("unexpected Go field name in Kafka JSON")
	}
	if _, exists := fields["schema_version"]; !exists {
		t.Fatal("missing schema_version")
	}
}

// TestKafkaProducerRejectsInvalidRecords 验证无效事件不能进入发送流程
func TestKafkaProducerRejectsInvalidRecords(t *testing.T) {
	producer := &kafkaProducer{topic: "wallet-test"}

	valid := walletEvent{
		SchemaVersion:  walletEventSchemaVersion,
		WalletGlobalID: uuid.New(),
	}

	tests := []struct {
		name      string
		partition int32
		event     walletEvent
	}{
		{
			name:      "negative partition",
			partition: -1,
			event:     valid,
		},
		{
			name:      "invalid schema version",
			partition: 0,
			event: walletEvent{
				SchemaVersion:  99,
				WalletGlobalID: valid.WalletGlobalID,
			},
		},
		{
			name:      "empty wallet global ID",
			partition: 0,
			event: walletEvent{
				SchemaVersion: walletEventSchemaVersion,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := producer.buildRecord(tt.partition, tt.event)
			if !errors.Is(err, ErrKafkaInvalidRecord) {
				t.Fatalf("expected ErrKafkaInvalidRecord, got %v", err)
			}
		})
	}
}
