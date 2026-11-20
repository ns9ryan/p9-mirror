package types

// ActivitySettings 活动配置（JSON 存储）
type ActivitySettings struct {
	DepositRewardType DepositRewardType `json:"deposit_reward_type"` // 充值奖励类型
	BetClaimMode      BetClaimMode      `json:"bet_claim_mode"`      // 投注奖励领取模式

	BonusMode     BonusMode      `json:"bonus_mode"`               // 奖励模式
	RewardSchemas []RewardSchema `json:"reward_schemas,omitempty"` // 奖励阶梯规则

	// 推荐活动专用
	ReferConfig *ReferConfig `json:"refer_config,omitempty"` // 推荐活动配置

	// 幸运转盘配置
	LuckWheel *LuckWheelConfig `json:"luck_wheel,omitempty"` // 幸运转盘配置
}

// LuckWheelConfig 幸运转盘配置
type LuckWheelConfig struct {
	// 幸运转盘：达标送抽奖机会；RewardSchemas 为奖池（Amount=奖金，Rate=权重）
	// WheelTriggerMode：1=仅投注、2=仅充值、3=投注|充值（任一达标即可）
	WheelTriggerMode WheelTriggerMode `json:"wheel_trigger_mode"`
	BetSchemas       []RewardSchema   `json:"bet_schemas"`     // Threshold=累计门槛，Amount=赠送次数（默认1）
	DepositSchemas   []RewardSchema   `json:"deposit_schemas"` // Threshold=累计门槛，Amount=赠送次数（默认1）
}

// RewardSchema 奖励阶梯规则
type RewardSchema struct {
	ID        int64 `json:"id"`
	Threshold int64 `json:"threshold"`
	Amount    int64 `json:"amount"`
	Rate      int   `json:"rate"`
	MaxBonus  int64 `json:"max_bonus"`
	MinBonus  int64 `json:"min_bonus"`
	Count     int   `json:"count"`
}

// LimitConfig 次数限制
type LimitConfig struct {
	PerUser  int `json:"per_user"`
	PerDay   int `json:"per_day"`
	PerWeek  int `json:"per_week"`
	PerMonth int `json:"per_month"`
	PerCycle int `json:"per_cycle"`
	Total    int `json:"total"`
}

// ReferConfig 推荐活动配置
type ReferConfig struct {
	MinDepositAmount int64 `json:"min_deposit_amount"`
	MinBetAmount     int64 `json:"min_bet_amount"`
	IPLimit          int   `json:"ip_limit"`
	DeviceLimit      int   `json:"device_limit"`
}
