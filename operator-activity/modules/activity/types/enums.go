package types

// ActivityType 活动类型
type ActivityType int

const (
	ActivityTypeDeposit        ActivityType = 1  // 充值
	ActivityTypeWagering       ActivityType = 2  // 洗码
	ActivityTypeReferralMember ActivityType = 3  // 推荐会员
	ActivityTypeLuckyWheel     ActivityType = 4  // 幸运转盘
	ActivityTypeAssist         ActivityType = 5  // 助力
	ActivityTypeBlindBox       ActivityType = 6  // 盲盒
	ActivityTypeNewbieBonus    ActivityType = 7  // 新人彩金
	ActivityTypeUniversalAgent ActivityType = 8  // 全民代理
	ActivityTypeRescueFund     ActivityType = 9  // 救援金
	ActivityTypeLuckyBet       ActivityType = 10 // 幸运注单
	ActivityTypeVIPBonus       ActivityType = 11 // VIP奖金
	ActivityTypeRedPacketRain  ActivityType = 12 // 红包雨
	ActivityTypeImageTextAd    ActivityType = 13 // 图文(广告)
	ActivityTypeCheckIn        ActivityType = 14 // 签到
)

// ActivityStatus 活动状态
type ActivityStatus int

const (
	ActivityStatusDraft   ActivityStatus = 1 // 草稿
	ActivityStatusOngoing ActivityStatus = 2 // 进行中
	ActivityStatusClosed  ActivityStatus = 3 // 已结束
)

// ClientType 客户端类型
type ClientType int

const (
	ClientTypeWeb     ClientType = 1 // PC网页
	ClientTypeMobile  ClientType = 2 // 移动网页
	ClientTypeAndroid ClientType = 3 // 安卓APP
	ClientTypeIOS     ClientType = 4 // iOS APP
)

// ApplyLimitType 参与活动限制类型
type ApplyLimitType int

const (
	ApplyLimitTypeUdid ApplyLimitType = 1 // 同设备号不限参与次数
	ApplyLimitTypeIP   ApplyLimitType = 2 // 同设备IP限参与一次
)

// ClaimLimitType 领取奖励限制类型
type ClaimLimitType int

const (
	ClaimLimitTypeDeposit  ClaimLimitType = 1 // 完成充值才能领取
	ClaimLimitTypeBank     ClaimLimitType = 2 // 完成银行卡绑定才能领取
	ClaimLimitTypeSms      ClaimLimitType = 3 // 完成短信验证才能领取
	ClaimLimitTypeEmail    ClaimLimitType = 4 // 完成邮箱验证才能领取
	ClaimLimitTypeRealname ClaimLimitType = 5 // 完成姓名填写才能领取
)

// DepositRewardType 充值奖励类型（可组合位掩码）
type DepositRewardType int

const (
	DepositFirst      DepositRewardType = 1 // 首充
	DepositSecond     DepositRewardType = 2 // 次存
	DepositCumulative DepositRewardType = 3 // 累计
	DepositSingle     DepositRewardType = 4 // 单笔
)

// CycleMode 循环模式
type CycleMode int

const (
	CycleUnspecified CycleMode = 0 // 未指定
	CycleNone        CycleMode = 1 // 不循环
	CycleDaily       CycleMode = 2 // 每日循环
	CycleWeekly      CycleMode = 3 // 每周循环
	CycleMonthly     CycleMode = 4 // 每月循环
	CycleInfinite    CycleMode = 5 // 无限循环
)

// ActivitySection 活动所属
type ActivitySection int

const (
	SectionLobby      ActivitySection = 1 // 大厅
	SectionElectronic ActivitySection = 2 // 电子
	SectionSports     ActivitySection = 3 // 体育
)

// ApplyMode 参与模式
type ApplyMode int

const (
	ApplyAuto   ApplyMode = 1 // 默认参与
	ApplyManual ApplyMode = 2 // 手动参与
)

// PayoutMode 派发奖励模式
type PayoutMode int

const (
	PayoutManualExpireVoid PayoutMode = 1 // 手动领取，过期作废
	PayoutManualExpireAuto PayoutMode = 2 // 手动领取，过期自动领取
	PayoutAuto             PayoutMode = 3 // 自动派发
)

// PayoutTiming 派发奖励时间
type PayoutTiming int

const (
	PayoutRealtime PayoutTiming = 1 // 实时
	PayoutNextDay  PayoutTiming = 2 // 次日
)

// BonusMode 奖金模式
type BonusMode int

const (
	BonusFixed            BonusMode = 1 // 固定金额多档
	BonusPercentage       BonusMode = 2 // 按比例多档
	BonusFixedTiered      BonusMode = 3 // 固定金额阶梯最高下发
	BonusPercentageTiered BonusMode = 4 // 按比例阶梯最高下发
	BonusSeparate         BonusMode = 5 // 单独奖励，如推荐活动，单独人数奖励
)

// OrderStatus 奖励订单状态
type OrderStatus int

const (
	OrderStatusPendingAudit  OrderStatus = 1 // 待审核
	OrderStatusApproved      OrderStatus = 2 // 审核通过
	OrderStatusRejected      OrderStatus = 3 // 审核拒绝
	OrderStatusPendingClaim  OrderStatus = 4 // 待领取
	OrderStatusClaimed       OrderStatus = 5 // 已领取
	OrderStatusExpired       OrderStatus = 6 // 已过期
	OrderStatusPendingPayout OrderStatus = 7 // 待派发
	OrderStatusPaidOut       OrderStatus = 8 // 已派发
)

// AuditStatus 审核状态
type AuditStatus int

const (
	AuditStatusManual AuditStatus = 1 // 人工审核
	AuditStatusRisk   AuditStatus = 2 // 风控拦截
)

// EventType 事件类型。具体取值由各活动插件定义。
type EventType int

// WageringBase 洗码计算基数
type WageringBase int

const (
	WageringBaseUnspecified       WageringBase = 0
	WageringBasePrincipalAndBonus WageringBase = 1 // 本金 + 赠送
	WageringBaseBonusOnly         WageringBase = 2 // 仅赠送
)

// BetClaimMode 投注奖励领取模式
type BetClaimMode int

const (
	BetClaimModeNone         BetClaimMode = 0
	BetClaimModeByItem       BetClaimMode = 1 // 逐条领取
	BetClaimModeByHighest    BetClaimMode = 2 // 最高档领取
	BetClaimModeByCumulative BetClaimMode = 3 // 累计领取
)

// WheelTriggerMode 幸运转盘达标触发维度（可位组合：仅投注=1、仅充值=2、两者皆可=1|2 位运算）
type WheelTriggerMode int

const (
	WheelTriggerUnspecified WheelTriggerMode = 0
	WheelTriggerBet         WheelTriggerMode = 1 // 累计有效投注
	WheelTriggerDeposit     WheelTriggerMode = 2 // 累计充值
)

// Has 是否包含指定触发维度；未配置时默认仅投注
func (m WheelTriggerMode) Has(flag WheelTriggerMode) bool {
	if m == WheelTriggerUnspecified {
		return flag == WheelTriggerBet
	}
	return m&flag != 0
}
