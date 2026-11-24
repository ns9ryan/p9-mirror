package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ErrInvalidWalletEvent 表示 Kafka 资金事件不合法
var ErrInvalidWalletEvent = errors.New("invalid wallet event")

// invalidWalletEvent 统一构造资金事件错误
func invalidWalletEvent(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidWalletEvent, reason)
}

// decodeWalletEvent 解码 Kafka 资金事件并校验消息完整性
func decodeWalletEvent(record *kgo.Record, router *router) (walletEvent, error) {
	if record == nil {
		return walletEvent{}, invalidWalletEvent("nil Kafka record")
	}

	if record.Topic == "" || record.Partition < 0 || record.Offset < 0 {
		return walletEvent{}, invalidWalletEvent("invalid Kafka record metadata")
	}

	if len(record.Value) == 0 {
		return walletEvent{}, invalidWalletEvent("empty Kafka record value")
	}

	var event walletEvent

	// 解析 Producer 写入的完整资金事件
	if err := json.Unmarshal(record.Value, &event); err != nil {
		return walletEvent{}, fmt.Errorf("%w: decode JSON: %w", ErrInvalidWalletEvent, err)
	}

	// 校验资金事件自身的数据完整性
	if err := event.validate(); err != nil {
		return walletEvent{}, err
	}

	// Kafka Key 必须与事件中的钱包全局标识一致
	if !bytes.Equal(record.Key, event.WalletGlobalID[:]) {
		return walletEvent{}, invalidWalletEvent("Kafka record key mismatch")
	}

	// 校验事件是否位于该钱包所属的固定分区
	partition, err := router.route(event.WalletGlobalID)
	if err != nil {
		return walletEvent{}, fmt.Errorf("%w: route wallet: %w", ErrInvalidWalletEvent, err)
	}

	if int(record.Partition) != partition {
		return walletEvent{}, fmt.Errorf("%w: partition mismatch: expected=%d actual=%d", ErrInvalidWalletEvent, partition, record.Partition)
	}

	return event, nil
}

// validate 校验资金事件的身份、业务来源、版本及余额变动一致性
func (e walletEvent) validate() error {
	if e.SchemaVersion != walletEventSchemaVersion {
		return fmt.Errorf("%w: unsupported schema version: %d", ErrInvalidWalletEvent, e.SchemaVersion)
	}

	if e.CreatedAt.IsZero() || e.CreatedAt.Nanosecond()%int(time.Millisecond) != 0 {
		return invalidWalletEvent("invalid event timestamp")
	}

	// 根据事件中的操作前快照还原钱包状态
	state := State{
		WalletID:         e.WalletID,                // 数据库钱包ID
		WalletGlobalID:   e.WalletGlobalID,          // 钱包全局唯一标识
		MemberGlobalID:   e.MemberGlobalID,          // 会员全局唯一标识
		OperatorCode:     e.OperatorCode,            // 所属厅业务编码
		CurrencyCode:     e.CurrencyCode,            // 钱包结算币种编码
		Balance:          e.Before.Balance,          // 操作前可用余额
		FrozenBalance:    e.Before.FrozenBalance,    // 操作前冻结余额
		UnsettledBalance: e.Before.UnsettledBalance, // 操作前未结算余额
		Version:          e.PreviousVersion,         // 操作前钱包版本
	}

	// 根据事件流水还原本次资金操作参数
	params := RecordParams{
		OperationID:    e.OperationID,                            // 钱包操作唯一标识
		BusinessType:   e.BusinessType,                           // 一级业务类型
		BusinessAction: e.BusinessAction,                         // 具体资金操作类型
		BusinessNo:     e.BusinessNo,                             // 原始业务单号
		Changes:        make([]BalanceChange, 0, len(e.Ledgers)), // 本次操作的余额变动集合
	}

	// 校验流水标识, 同时提取每种余额的变动金额
	ledgerIDs := make(map[uuid.UUID]struct{}, len(e.Ledgers))

	for _, ledger := range e.Ledgers {
		if ledger.LedgerGlobalID == uuid.Nil {
			return invalidWalletEvent("empty ledger global ID")
		}

		if _, exists := ledgerIDs[ledger.LedgerGlobalID]; exists {
			return invalidWalletEvent("duplicate ledger global ID")
		}
		ledgerIDs[ledger.LedgerGlobalID] = struct{}{}

		params.Changes = append(params.Changes, BalanceChange{
			BalanceType:  ledger.BalanceType,  // 余额类型
			ChangeAmount: ledger.ChangeAmount, // 变动金额
		})
	}

	// 复用生产端资金校验及计算规则, 仅验证事件结果是否一致
	prepared, err := prepareRecord(state, params)
	if err != nil {
		return fmt.Errorf("%w: validate balance changes: %w", ErrInvalidWalletEvent, err)
	}

	// 一次资金操作必须且只能递增一次钱包版本
	if prepared.nextState.Version != e.Version {
		return invalidWalletEvent("wallet version mismatch")
	}

	// 操作后全部余额必须与流水计算结果一致
	if balancesFromState(prepared.nextState) != e.After {
		return invalidWalletEvent("wallet balance snapshot mismatch")
	}

	// 逐条校验流水中的变动前后余额
	for i, ledger := range e.Ledgers {
		expected := prepared.ledgers[i]

		if ledger.BalanceBefore != expected.BalanceBefore || ledger.BalanceAfter != expected.BalanceAfter {
			return fmt.Errorf("%w: ledger balance mismatch: balance_type=%d", ErrInvalidWalletEvent, ledger.BalanceType)
		}
	}

	return nil
}
