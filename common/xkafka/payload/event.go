package payload

import "time"

type SingleMemberEvent struct {
	EventType    string    `json:"event_type"`    // 事件类型
	OperatorCode string    `json:"operator_code"` // 分站标识，必填
	MemberID     int64     `json:"member_id"`     // 会员ID，可选
	OccurredAt   time.Time `json:"occurred_at"`   // 业务事实发生时间
	PublishedAt  time.Time `json:"published_at"`  // 事件发布时间
	TraceID      string    `json:"trace_id"`      // 链路追踪ID，可选
}
