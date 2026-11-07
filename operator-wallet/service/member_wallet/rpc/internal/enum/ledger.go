package enum

// BalanceType 会员钱包余额类型
type BalanceType int64

const (
	BalanceTypeAvailable BalanceType = 1 // 可用余额
	BalanceTypeFrozen    BalanceType = 2 // 冻结余额
	BalanceTypeUnsettled BalanceType = 3 // 游戏未结算余额
)

// IsValid 判断余额类型是否有效
func (t BalanceType) IsValid() bool {
	switch t {
	case BalanceTypeAvailable,
		BalanceTypeFrozen,
		BalanceTypeUnsettled:
		return true
	default:
		return false
	}
}

// BusinessType 资金业务主类型
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

// IsValid 判断业务主类型是否有效
func (t BusinessType) IsValid() bool {
	switch t {
	case BusinessTypeDeposit,
		BusinessTypeWithdraw,
		BusinessTypeGame,
		BusinessTypeActivity,
		BusinessTypeTask,
		BusinessTypeCommission,
		BusinessTypeSystemReward:
		return true
	default:
		return false
	}
}

// BusinessAction 具体资金操作类型, 按业务主类型分段编码
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

// IsValidFor 判断操作类型是否属于指定业务主类型
func (a BusinessAction) IsValidFor(businessType BusinessType) bool {
	switch businessType {
	// 存款
	case BusinessTypeDeposit:
		switch a {
		case BusinessActionDepositCredit, BusinessActionDepositBonus:
			return true
		}

	// 提现
	case BusinessTypeWithdraw:
		switch a {
		case BusinessActionWithdrawalFreeze,
			BusinessActionWithdrawalPayout,
			BusinessActionWithdrawalFee,
			BusinessActionWithdrawalUnfreeze:
			return true
		}

	// 游戏
	case BusinessTypeGame:
		switch a {
		case BusinessActionGameBet,
			BusinessActionGameSettlement,
			BusinessActionGameBetCancel,
			BusinessActionGameTip,
			BusinessActionGameRebate,
			BusinessActionGameSettlementRollback:
			return true
		}

	// 活动
	case BusinessTypeActivity:
		return a == BusinessActionActivityReward

	// 任务
	case BusinessTypeTask:
		return a == BusinessActionTaskReward

	// 佣金
	case BusinessTypeCommission:
		return a == BusinessActionCommissionCredit

	// 系统奖励
	case BusinessTypeSystemReward:
		switch a {
		case BusinessActionFeedbackReward, BusinessActionVIPReward:
			return true
		}
	}

	return false
}

// TODO: 后续接入新业务时, 按所属 BusinessType 补充资金操作类型
