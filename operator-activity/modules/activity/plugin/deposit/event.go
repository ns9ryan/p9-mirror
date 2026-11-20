package deposit

import (
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

const EventApproved types.EventType = 2

// Event 充值活动事件
type Event struct {
	OperatorCode string
	MemberID     int64
	OrderID      int64
	Amount       int64
	DepositCount int
	ClientType   int
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

// GetAmount 获取充值金额
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
func (e *Event) GetGameID() int64 { return 0 }

// GetEventTime 获取事件时间
func (e *Event) GetEventTime() time.Time {
	if e == nil {
		return time.Time{}
	}
	return e.EventTime
}

// GetEventType 获取事件类型
func (e *Event) GetEventType() types.EventType { return EventApproved }

// GetDepositCount 获取充值次数
func (e *Event) GetDepositCount() int {
	if e == nil {
		return 0
	}
	return e.DepositCount
}
