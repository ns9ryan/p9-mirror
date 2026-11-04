package enum

// BalanceType 会员钱包余额类型
type BalanceType int64

const (
	BalanceTypeAvailable BalanceType = 1 // 可用余额
	BalanceTypeFrozen    BalanceType = 2 // 冻结余额
	BalanceTypeUnsettled BalanceType = 3 // 游戏未结算余额
)

// BusinessType 资金业务类型
type BusinessType int64

const (
	BusinessTypeDeposit BusinessType = 1 // 存款
)

// BusinessAction 具体资金操作类型
type BusinessAction int64

const (
	BusinessActionDepositCredit BusinessAction = 1001 // 存款本金入账
)
