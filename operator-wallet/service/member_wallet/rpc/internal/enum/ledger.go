package enum

// BalanceType 会员钱包余额类型
type BalanceType int64

const (
	BalanceTypeAvailable BalanceType = 1 // 可用余额
	BalanceTypeFrozen    BalanceType = 2 // 冻结余额
	BalanceTypeUnsettled BalanceType = 3 // 游戏未结算余额
)

// BusinessType 资金业务主类型 (1: 存款, 2: 提现, 3: 游戏, 4: 活动, 5: 任务, 6: 佣金, 7: 系统奖励)
type BusinessType int64

const (
	BusinessTypeDeposit      BusinessType = 1 // 存款
	BusinessTypeWithdraw     BusinessType = 2 // 提现
	BusinessTypeGame         BusinessType = 3 // 游戏
	BusinessTypeActivity     BusinessType = 4 // 活动
	BusinessTypeTask         BusinessType = 5 // 任务
	BusinessTypeCommission   BusinessType = 6 // 佣金
	BusinessTypeSystemReward BusinessType = 7 // 系统奖励
)

// BusinessAction 具体资金操作类型
//
// 编码按 BusinessType 分段:
// 1000-1999: 存款
// 2000-2999: 提现
// 3000-3999: 游戏
// 4000-4999: 活动
// 5000-5999: 任务
// 6000-6999: 佣金
// 7000-7999: 系统奖励
type BusinessAction int64

const (
	// ---------------------------------------------------------------------
	// 存款
	// ---------------------------------------------------------------------

	BusinessActionDepositCredit BusinessAction = 1001 // 存款本金入账
	BusinessActionDepositBonus  BusinessAction = 1002 // 存款赠送入账

	// ---------------------------------------------------------------------
	// 提现
	// ---------------------------------------------------------------------

	BusinessActionWithdrawalFreeze   BusinessAction = 2001 // 提现冻结
	BusinessActionWithdrawalPayout   BusinessAction = 2002 // 提现出款
	BusinessActionWithdrawalFee      BusinessAction = 2003 // 提现手续费
	BusinessActionWithdrawalUnfreeze BusinessAction = 2004 // 提现解冻退回

	// ---------------------------------------------------------------------
	// 游戏
	// ---------------------------------------------------------------------

	BusinessActionGameBet                BusinessAction = 3001 // 游戏下注
	BusinessActionGameSettlement         BusinessAction = 3002 // 游戏结算
	BusinessActionGameBetCancel          BusinessAction = 3003 // 下注取消
	BusinessActionGameTip                BusinessAction = 3004 // 游戏消费打赏
	BusinessActionGameRebate             BusinessAction = 3005 // 游戏返水/返利
	BusinessActionGameSettlementRollback BusinessAction = 3006 // 结算后回滚

	// ---------------------------------------------------------------------
	// 活动
	// ---------------------------------------------------------------------

	BusinessActionActivityReward BusinessAction = 4001 // 活动奖励入账

	// ---------------------------------------------------------------------
	// 任务
	// ---------------------------------------------------------------------

	BusinessActionTaskReward BusinessAction = 5001 // 任务奖励入账

	// ---------------------------------------------------------------------
	// 佣金
	// ---------------------------------------------------------------------

	BusinessActionCommissionCredit BusinessAction = 6001 // 全民代理佣金入账

	// ---------------------------------------------------------------------
	// 系统奖励
	// ---------------------------------------------------------------------

	BusinessActionFeedbackReward BusinessAction = 7001 // 有奖反馈奖励入账
	BusinessActionVIPReward      BusinessAction = 7002 // VIP 奖励入账
)

// TODO: 后续接入新业务时, 按所属 BusinessType 补充资金操作类型
