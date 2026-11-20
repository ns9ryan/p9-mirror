package types

import "time"

// Activity 活动
type Activity struct {
	ID                int64
	OperatorCode      string
	Type              ActivityType     // 活动类型
	Status            ActivityStatus   // 活动状态
	ApplyMode         ApplyMode        // 参与活动模式
	CycleMode         CycleMode        // 循环模式
	AllowedFixedLevel []int64          // 固定等级
	AllowedAutoLevel  []int64          // 自动等级
	AllowedClients    []int64          // 客户端类型
	AllowedGameIds    []int64          // 赠送奖励指定游戏洗码
	AllowedAgentIds   []int64          // 指定代理ID
	ApplyLimit        []ApplyLimitType // 参与活动限制
	ClaimLimit        []ClaimLimitType // 领取奖励限制
	OpenAt            *time.Time       // 活动开始时间
	CloseAt           *time.Time       // 活动结束时间
	IsPermanent       bool             // 活动是否永久，为 true 时可不依赖结束时间
	ReleaseAt         *time.Time       // 活动发布时间
	AutoAudit         bool             // 是否开启自动审核；关闭则全部走人工审核
	AutoAuditMaxAmt   int64            // 自动审核金额上限：RewardsTotal < 该值自动通过，否则人工审核
	PayoutMode        PayoutMode       // 派发奖励模式
	PayoutTiming      PayoutTiming     // 派发奖励时间
	ClaimExpireSec    int64            // 手动领取过期秒数，0 表示不过期
	Settings          ActivitySettings // 活动配置
	Version           int
	UpdatedAt         time.Time
	CreatedAt         time.Time
}

// IsEnd 活动是否结束（IsPermanent 为 false 且 eventTime不在openAt和CloseAt之间）
func (p *Activity) IsEnd(eventTime time.Time) bool {
	if p.IsPermanent {
		return false
	}
	if p.OpenAt != nil && !p.OpenAt.IsZero() && eventTime.Before(*p.OpenAt) {
		return false
	}
	if p.CloseAt != nil && !p.CloseAt.IsZero() && eventTime.After(*p.CloseAt) {
		return true
	}
	return false
}

// IsRelease 活动是否发布（ReleaseAt 非空且非零值）
func (p *Activity) IsRelease() bool {
	return p.ReleaseAt != nil && !p.ReleaseAt.IsZero()
}

// OpBaseInfo 分站基础信息
type OpBaseInfo struct {
	Currency string // 货币单位
	TimeZone string // 时区
}

// MemberInfo 会员信息
type MemberInfo struct {
	ID           int64   // 会员ID
	OperatorCode string  // 分站编码
	FixedLevels  []int64 // 固定等级
	AutoLevel    int64   // 自动等级
	// ClientType   int     // 客户端类型
	InviterID    int64  // 推荐人ID
	AgentID      int64  // 代理ID
	RegisterIP   string // 注册IP
	RegisterUdid string // 注册设备码
}

// ActivityEvent 各插件事件对流水线暴露的公共数据。插件没有的字段返回零值。
type ActivityEvent interface {
	GetOperatorCode() string
	GetMemberID() int64
	GetOrderID() int64
	GetAmount() int64
	GetClientType() int
	GetGameID() int64
	GetEventTime() time.Time
	GetEventType() EventType
	GetDepositCount() int
}

// TriggerMetric 触发指标（由 metric 策略计算）
type TriggerMetric struct {
	DepositCount          int
	DepositAmount         int64
	CumulativeDeposit     int64
	BetAmount             int64
	BetCount              int64
	CumulativeBet         int64
	TotalReferralCount    int
	ValidReferralCount    int
	NewVaildReferralCount int
	FromDeposit           bool // 转盘本次按累计充值判档
}

// RewardItem 计算出的奖励项
type RewardItem struct {
	SchemaID       int64
	Amount         int64
	Remark         string
	IdempotencyKey string
}

// ActivityOrder 奖励订单
type ActivityOrder struct {
	ID            int64
	OrderNo       string
	OperatorCode  string
	MemberID      int64
	AgentID       int64
	ActivityID    int64
	ActivityType  ActivityType
	Round         int
	EventType     EventType
	RelOrderID    int64
	RelAmount     int64
	Status        OrderStatus
	AuditStatus   AuditStatus
	ApprovedAt    time.Time
	ApproverUID   int64
	RewardsTotal  int64
	PayoutMode    PayoutMode
	PayoutTiming  PayoutTiming
	ClientType    ClientType
	PayoutAt      time.Time
	ClaimExp      time.Time
	Remark        string
	Version       int
	SubOrderCount int
	AppliedAt     time.Time
	CreatedAt     time.Time
	Items         []*ActivityOrderItem
}

// PromotionOrderItem 奖励订单项
type ActivityOrderItem struct {
	ID             int64
	OperatorCode   string
	MemberID       int64
	OrderID        int64
	ActivityID     int64
	ActivityType   ActivityType
	Round          int
	RewardAmount   int64
	IdempotencyKey string
	CreatedAt      time.Time
}

// TimeWindow 循环周期时间窗
type TimeWindow struct {
	Start time.Time
	End   time.Time
}
