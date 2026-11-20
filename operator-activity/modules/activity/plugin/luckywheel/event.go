package luckywheel

import (
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

const (
	EventBet     types.EventType = 3
	EventDeposit types.EventType = 2
	EventSpin    types.EventType = 4
)

// Event 转盘活动事件，覆盖投注、充值与抽奖。
type Event struct {
	EventType    types.EventType
	OperatorCode string
	MemberID     int64
	OrderID      int64
	Amount       int64
	ClientType   int
	GameID       int64
	EventTime    time.Time
}

// GetOperatorCode 获取分站编码
func (e *Event) GetOperatorCode() string {
	if e == nil {
		return ""
	}
	return e.OperatorCode
}

// GetMemberID 获取会员ID
func (e *Event) GetMemberID() int64 {
	if e == nil {
		return 0
	}
	return e.MemberID
}

// GetOrderID 获取订单ID
func (e *Event) GetOrderID() int64 {
	if e == nil {
		return 0
	}
	return e.OrderID
}

// GetAmount 获取投注金额
func (e *Event) GetAmount() int64 {
	if e == nil {
		return 0
	}
	return e.Amount
}

// GetClientType 获取客户端类型
func (e *Event) GetClientType() int {
	if e == nil {
		return 0
	}
	return e.ClientType
}

// GetGameID 获取游戏ID
func (e *Event) GetGameID() int64 {
	if e == nil {
		return 0
	}
	return e.GameID
}

// GetEventTime 获取事件时间
func (e *Event) GetEventTime() time.Time {
	if e == nil {
		return time.Time{}
	}
	return e.EventTime
}

// GetEventType 获取事件类型
func (e *Event) GetEventType() types.EventType {
	if e == nil {
		return 0
	}
	return e.EventType
}

// GetDepositCount 获取投注次数
func (e *Event) GetDepositCount() int { return 0 }
