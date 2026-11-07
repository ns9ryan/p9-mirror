package wallet

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// 统一记账业务错误
var (
	ErrInvalidRecordParams = errors.New("invalid wallet record parameters") // 记账参数不合法
	ErrInsufficientBalance = errors.New("insufficient wallet balance")      // 钱包余额不足
	ErrBalanceOverflow     = errors.New("wallet balance overflow")          // 钱包余额溢出
	ErrOperationConflict   = errors.New("wallet operation conflict")        // 钱包操作标识冲突
)

// invalidRecordParams 统一构造记账参数错误
func invalidRecordParams(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRecordParams, reason)
}

// BalanceChange 单种钱包余额变动
type BalanceChange struct {
	BalanceType  enum.BalanceType // 余额类型
	ChangeAmount int64            // 变动金额, 正数增加, 负数减少
}

// validate 校验单种余额变动
func (c BalanceChange) validate() error {
	if !c.BalanceType.IsValid() {
		return invalidRecordParams("unsupported balance type")
	}

	if c.ChangeAmount == 0 {
		return invalidRecordParams("zero balance change")
	}

	return nil
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

// validate 校验统一记账参数
func (p RecordParams) validate() error {
	// 校验操作标识
	if p.OperationID == uuid.Nil {
		return invalidRecordParams("empty operation ID")
	}

	// 判断业务主类型是否有效
	if !p.BusinessType.IsValid() {
		return invalidRecordParams("unsupported business type")
	}

	// 判断操作类型是否属于指定业务主类型
	if !p.BusinessAction.IsValidFor(p.BusinessType) {
		return invalidRecordParams("business action does not match business type")
	}

	if p.BusinessNo == "" || strings.TrimSpace(p.BusinessNo) != p.BusinessNo || utf8.RuneCountInString(p.BusinessNo) > 128 {
		return invalidRecordParams("invalid business number")
	}

	if len(p.Changes) == 0 {
		return invalidRecordParams("empty balance changes")
	}

	// 同一种余额在一次操作中只能变动一次
	seen := make(map[enum.BalanceType]struct{}, len(p.Changes))

	for _, change := range p.Changes {
		if err := change.validate(); err != nil {
			return err
		}

		if _, exists := seen[change.BalanceType]; exists {
			return invalidRecordParams("duplicate balance type")
		}
		seen[change.BalanceType] = struct{}{}
	}

	return nil
}

// applyBalanceChange 计算余额变动并检查透支和溢出
func applyBalanceChange(balance, change int64, balanceType enum.BalanceType) (int64, error) {
	// 增加余额时检查 int64 上限
	if change > 0 && balance > math.MaxInt64-change {
		return 0, fmt.Errorf(
			"%w: balance_type=%d, balance=%d, change=%d",
			ErrBalanceOverflow, balanceType, balance, change,
		)
	}

	// 负数直接比较, 避免对 math.MinInt64 取反造成溢出
	if change < 0 && change < -balance {
		return 0, fmt.Errorf(
			"%w: balance_type=%d, balance=%d, change=%d",
			ErrInsufficientBalance, balanceType, balance, change,
		)
	}

	return balance + change, nil
}
