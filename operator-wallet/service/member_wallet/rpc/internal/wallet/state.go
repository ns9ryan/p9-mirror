// ============================================================================
// 钱包运行时状态及校验
// ============================================================================

package wallet

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ErrInvalidWalletState 表示钱包状态不合法
var ErrInvalidWalletState = errors.New("invalid wallet state")

// State 表示会员钱包在 Wallet Engine 中的运行时状态
type State struct {
	// 钱包标识
	WalletID       int64     // 数据库钱包ID
	WalletGlobalID uuid.UUID // 钱包全局唯一标识
	MemberGlobalID uuid.UUID // 会员全局唯一标识

	// 归属信息
	OperatorCode string // 所属厅业务编码
	CurrencyCode string // 钱包结算币种编码

	// 余额
	Balance          int64 // 可用余额
	FrozenBalance    int64 // 冻结余额
	UnsettledBalance int64 // 游戏未结算余额

	Version int64 // 钱包状态版本, 随资金操作递增
}

// Validate 校验钱包运行时状态
func (s State) Validate() error {
	switch {
	// 校验钱包标识
	case s.WalletID <= 0:
		return invalidWalletState("invalid wallet ID")

	case s.WalletGlobalID == uuid.Nil:
		return invalidWalletState("empty wallet global ID")

	case s.MemberGlobalID == uuid.Nil:
		return invalidWalletState("empty member global ID")

	// 校验归属信息
	case s.OperatorCode == "" ||
		strings.TrimSpace(s.OperatorCode) != s.OperatorCode ||
		utf8.RuneCountInString(s.OperatorCode) > 64:
		return invalidWalletState("invalid operator code")

	case s.CurrencyCode == "" ||
		strings.TrimSpace(s.CurrencyCode) != s.CurrencyCode ||
		utf8.RuneCountInString(s.CurrencyCode) > 16:
		return invalidWalletState("invalid currency code")

	// 校验余额
	case s.Balance < 0 || s.FrozenBalance < 0 || s.UnsettledBalance < 0:
		return invalidWalletState("negative wallet balance")

	// 校验版本
	case s.Version < 0:
		return invalidWalletState("negative wallet version")
	}

	return nil
}

// invalidWalletState 统一构造钱包状态错误
func invalidWalletState(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidWalletState, reason)
}
