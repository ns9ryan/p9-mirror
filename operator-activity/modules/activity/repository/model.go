package repository

import (
	"encoding/json"
	"time"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// derefTime 解引用时间
func derefTime(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}

// ActivityModel 活动表 DTO
type ActivityModel struct {
	ID                 int64
	OperatorCode       string
	Type               types.ActivityType   // 活动类型
	Status             types.ActivityStatus // 活动状态
	ApplyMode          types.ApplyMode      // 参与活动模式
	CycleMode          types.CycleMode      // 循环模式
	AllowedFixedLevel  []int64              // 固定等级
	AllowedAutoLevel   []int64              // 自动等级
	AllowedClients     []int64              // 客户端类型
	AllowedGameIds     []int64              // 指定游戏ID
	AllowedAgentIds    []int64              // 指定代理ID
	ApplyLimit         []int64              // 参与活动限制
	ClaimLimit         []int64              // 领取奖励限制
	IsPermanent        bool                 // 活动是否永久，为 true 时可不依赖结束时间
	OpenAt             *time.Time
	CloseAt            *time.Time
	ReleaseAt          *time.Time
	AutoAudit          bool                  // 是否开启自动审核；关闭则全部走人工审核
	AutoAuditMaxAmt    int64                 // 自动审核金额上限：RewardsTotal < 该值自动通过，否则人工审核
	WageringMultiplier int                   // 洗码倍数，如 1 表示 1 倍
	WageringBase       types.WageringBase    // 洗码计算基数: 1-本金+赠送 2-仅赠送
	PayoutMode         types.PayoutMode      // 派发奖励模式
	PayoutTiming       types.PayoutTiming    // 派发奖励时间
	Section            types.ActivitySection // 活动所属: 1大厅 2电子 3体育
	ClaimExpireSec     int64                 // 手动领取过期秒数，0 表示不过期
	Settings           string
	ShowInList         bool
	ShowInSidebar      bool
	CreateUid          int64
	UpdateUid          int64
	Version            int
	SortWeight         int
	UpdatedAt          time.Time
	CreatedAt          time.Time
}

// ApplicationModel 参与记录 DTO
type ApplicationModel struct {
	ID          int64
	PartnerCode string
	Mode        int
	PromoID     int64
	UserID      int64
	ApplyMode   int
	ApplyAt     time.Time
	ApplyIP     string
	ApplyFP     string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

// ReferModel 活动推荐关系 DTO
type ReferModel struct {
	ID          int64
	PartnerCode string
	InviterID   int64
	UserID      int64
	PromoID     int64
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

// WheelChanceBalanceModel 转盘机会余额 DTO
type WheelChanceBalanceModel struct {
	ID           int64
	OperatorCode string
	ActivityID   int64
	MemberID     int64
	Granted      int
	Used         int
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// WheelChanceGrantModel 转盘机会发放流水 DTO
type WheelChanceGrantModel struct {
	ID             int64
	OperatorCode   string
	ActivityID     int64
	MemberID       int64
	Round          int
	SchemaID       int64
	EventType      types.EventType
	OrderID        int64
	Chances        int
	IdempotencyKey string
	CreatedAt      time.Time
}

// PromoStatModel 活动领取统计 DTO
type ActivityStatModel struct {
	ID                int64
	OperatorCode      string
	ActivityID        int64
	ClaimUserCount    int64
	ClaimedAmount     int64
	ClaimedOrderCount int64
	UpdatedAt         time.Time
	CreatedAt         time.Time
}

// PromoStatClaimerModel 领取用户去重 DTO
type ActivityStatClaimerModel struct {
	ID           int64
	OperatorCode string
	ActivityID   int64
	MemberID     int64
	CreatedAt    time.Time
}

// PromoStatOrderModel 已计入统计的订单 DTO
type ActivityStatOrderModel struct {
	ID           int64
	OrderID      int64
	OperatorCode string
	ActivityID   int64
	MemberID     int64
	Amount       int64
	CreatedAt    time.Time
}

// PromoOrderModel 奖励订单 DTO
type ActivityOrderModel struct {
	ID            int64
	OrderNo       string
	OperatorCode  string
	MemberID      int64
	AgentID       int64
	ActivityID    int64
	ActivityType  types.ActivityType
	Round         int
	EventType     types.EventType
	RelOrderID    int64
	RelAmount     int64
	Status        int
	AuditStatus   int
	ApprovedAt    *time.Time
	ApproverUID   int64
	RewardsTotal  int64
	PayoutMode    int
	PayoutTiming  int
	PayoutAt      *time.Time
	ClaimExp      *time.Time
	Remark        string
	Version       int
	SubOrderCount int
	AppliedAt     time.Time
	CreatedAt     time.Time
}

// PromoOrderItemModel 奖励订单项 DTO
type ActivityOrderItemModel struct {
	ID             int64
	OrderID        int64
	OperatorCode   string
	MemberID       int64
	ActivityID     int64
	ActivityType   types.ActivityType
	Round          int
	RewardAmount   int64
	IdempotencyKey string
	CreatedAt      time.Time
}

// PayoutScheduleModel 延迟派发队列 DTO
type PayoutScheduleModel struct {
	OrderID   int64
	PayoutAt  time.Time
	Status    int
	CreatedAt time.Time
}

// DepositOrderModel 充值订单 DTO
type DepositOrderModel struct {
	ID            int64
	OperatorCode  string
	MemberID      int64
	DepositAmount int64
	Status        int
	CreatedAt     time.Time
}

// BetOrderModel 投注订单 DTO
type BetOrderModel struct {
	ID           int64
	OperatorCode string
	MemberID     int64
	BetAmount    int64
	GameID       int64
	CreatedAt    time.Time
}

// UserModel 会员 DTO
type UserModel struct {
	ID           int64
	OperatorCode string
	FixedLevels  []int64
	AutoLevel    int64
	ClientType   int
	InviterID    int64
	AgentID      int64
	Balance      int64
	Username     string
	Password     string
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

func (m *ActivityModel) ToActivity() (*types.Activity, error) {
	return toActivity(m)
}

func ActivityFromEnt(p *ent.Activity) *ActivityModel {
	if p == nil {
		return nil
	}
	return &ActivityModel{
		ID:                 p.ID,
		OperatorCode:       p.OperatorCode,
		Type:               types.ActivityType(p.Type),
		Status:             types.ActivityStatus(p.Status),
		ApplyMode:          types.ApplyMode(p.ApplyMode),
		CycleMode:          types.CycleMode(p.CycleMode),
		IsPermanent:        p.IsPermanent,
		OpenAt:             p.OpenAt,
		CloseAt:            p.CloseAt,
		ReleaseAt:          p.ReleaseAt,
		AutoAudit:          p.AutoAudit,
		AutoAuditMaxAmt:    p.AutoAuditMaxAmt,
		WageringMultiplier: p.WageringMultiplier,
		WageringBase:       types.WageringBase(p.WageringBase),
		PayoutMode:         types.PayoutMode(p.PayoutMode),
		PayoutTiming:       types.PayoutTiming(p.PayoutTiming),
		Section:            types.ActivitySection(p.Section),
		ClaimExpireSec:     p.ClaimExpireSec,
		Settings:           p.Settings,
		AllowedFixedLevel:  p.AllowedFixedLevel,
		AllowedAutoLevel:   p.AllowedAutoLevel,
		AllowedClients:     p.AllowedClients,
		AllowedGameIds:     p.AllowedGameIds,
		AllowedAgentIds:    p.AllowedAgentIds,
		ApplyLimit:         p.ApplyLimit,
		ClaimLimit:         p.ClaimLimit,
		ShowInList:         p.ShowInList,
		ShowInSidebar:      p.ShowInSidebar,
		CreateUid:          p.CreateUID,
		UpdateUid:          p.UpdateUID,
		Version:            p.Version,
		SortWeight:         p.SortWeight,
		UpdatedAt:          p.UpdatedAt,
		CreatedAt:          p.CreatedAt,
	}
}

func activityOrderFromEnt(m *ent.ActivityOrder) *ActivityOrderModel {
	if m == nil {
		return nil
	}
	return &ActivityOrderModel{
		ID:            m.ID,
		OrderNo:       m.OrderNo,
		OperatorCode:  m.OperatorCode,
		MemberID:      m.MemberID,
		AgentID:       m.AgentID,
		ActivityID:    m.ActivityID,
		ActivityType:  types.ActivityType(m.Type),
		Round:         m.Round,
		EventType:     types.EventType(m.EventType),
		RelOrderID:    m.RelEventOrderID,
		RelAmount:     m.RelAmount,
		Status:        m.Status,
		AuditStatus:   m.AuditStatus,
		ApprovedAt:    m.ApprovedAt,
		ApproverUID:   m.ApproverUID,
		RewardsTotal:  m.RewardsTotal,
		PayoutMode:    m.PayoutMode,
		PayoutTiming:  m.PayoutTiming,
		PayoutAt:      m.PayoutAt,
		ClaimExp:      m.ClaimExpireAt,
		Remark:        m.Remark,
		Version:       m.Version,
		SubOrderCount: m.SubOrderCount,
		AppliedAt:     derefTime(m.AppliedAt),
		CreatedAt:     m.CreatedAt,
	}
}

func toActivity(m *ActivityModel) (*types.Activity, error) {
	var settings types.ActivitySettings
	if m.Settings != "" {
		if err := json.Unmarshal([]byte(m.Settings), &settings); err != nil {
			return nil, err
		}
	}
	return &types.Activity{
		ID:                m.ID,
		OperatorCode:      m.OperatorCode,
		Type:              types.ActivityType(m.Type),
		Status:            types.ActivityStatus(m.Status),
		ApplyMode:         types.ApplyMode(m.ApplyMode),
		CycleMode:         types.CycleMode(m.CycleMode),
		AllowedFixedLevel: m.AllowedFixedLevel,
		AllowedAutoLevel:  m.AllowedAutoLevel,
		AllowedClients:    m.AllowedClients,
		AllowedGameIds:    m.AllowedGameIds,
		AllowedAgentIds:   m.AllowedAgentIds,
		ApplyLimit:        types.Int64sToApplyLimitTypes(m.ApplyLimit),
		ClaimLimit:        types.Int64sToClaimLimitTypes(m.ClaimLimit),
		IsPermanent:       m.IsPermanent,
		OpenAt:            m.OpenAt,
		CloseAt:           m.CloseAt,
		ReleaseAt:         m.ReleaseAt,
		AutoAudit:         m.AutoAudit,
		AutoAuditMaxAmt:   m.AutoAuditMaxAmt,
		PayoutMode:        m.PayoutMode,
		PayoutTiming:      m.PayoutTiming,
		ClaimExpireSec:    m.ClaimExpireSec,
		Settings:          settings,
		Version:           m.Version,
		UpdatedAt:         m.UpdatedAt,
		CreatedAt:         m.CreatedAt,
	}, nil
}
