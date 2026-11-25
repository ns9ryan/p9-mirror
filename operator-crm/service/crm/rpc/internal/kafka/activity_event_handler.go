package kafka

import (
	"context"
	"time"

	"oa.98ent.com/p9/common/xkafka/topic"
)

// depositPayload 对应 operator-activity 侧的 types.DepositPayload，
// 字段顺序与 json tag 须与消费端保持一致。
type depositPayload struct {
	OperatorCode   string    `json:"operator_code"`    // 厅主编码，用于按厅主隔离消费
	MemberID       int64     `json:"member_id"`        // 会员ID
	OrderID        int64     `json:"order_id"`         // 存款订单ID
	Amount         int64     `json:"amount"`           // 充值金额（分）
	ClientType     int       `json:"client_type"`      // 客户端类型，0=未知（暂固定）
	Currency       string    `json:"currency"`         // 充值币种，如 USDT
	Timestamp      int64     `json:"timestamp"`        // 事件时间 unix 毫秒
	DepositCount   int       `json:"deposit_count"`    // 会员累计充值次数
	EventType      string    `json:"event_type"`       // 兼容充值任务 事件类型 和topic一样
	OccurredAt     time.Time `json:"occurred_at"`      // 兼容充值任务 业务事实发生时间
	PublishedAt    time.Time `json:"published_at"`     // 兼容充值任务 事件发布时间
	TraceID        string    `json:"trace_id"`         // 兼容充值任务 链路追踪ID，可选
	MemberGlobalID string    `json:"member_global_id"` //兼容充值任务 会员全局 ID
}

// betPayload 对应 operator-activity 侧的 types.BetPayload。
type betPayload struct {
	OperatorCode   string    `json:"operator_code"`    // 厅主编码，用于按厅主隔离消费
	MemberID       int64     `json:"member_id"`        // 会员ID
	OrderID        int64     `json:"order_id"`         // 流水记录ID（Turnover.ID）
	Amount         int64     `json:"amount"`           // 消费金额（分），正数=赢，负数=输，0=模拟
	GameID         int64     `json:"game_id"`          // 游戏ID，0=未关联具体游戏
	ClientType     int       `json:"client_type"`      // 客户端类型，0=未知
	Currency       string    `json:"currency"`         // 结算币种，如 USDT
	Timestamp      int64     `json:"timestamp"`        // 事件时间 unix 毫秒
	EventType      string    `json:"event_type"`       // 兼容活动任务 事件类型，与 topic 一致
	OccurredAt     time.Time `json:"occurred_at"`      // 兼容活动任务 业务事实发生时间
	PublishedAt    time.Time `json:"published_at"`     // 兼容活动任务 事件发布时间
	TraceID        string    `json:"trace_id"`         // 兼容活动任务 链路追踪ID，可选
	MemberGlobalID string    `json:"member_global_id"` // 会员全局 ID
}

// DepositApprovedEventInput 存款审核通过事件的输入参数。
// memberGlobalID 由调用方提前查好后传入，避免在此重复查询。
type DepositApprovedEventInput struct {
	OperatorCode   string
	MemberID       int64
	OrderID        int64
	Amount         int64
	Currency       string
	DepositCount   int
	ClientType     int
	MemberGlobalID string
}

// GameOrderSettledEventInput 游戏订单结算事件的输入参数。
// memberGlobalID 由调用方提前查好后传入。
type GameOrderSettledEventInput struct {
	OperatorCode   string
	MemberID       int64
	OrderID        int64
	GameID         int64
	Amount         int64
	Currency       string
	ClientType     int
	MemberGlobalID string
}

// PublishDepositOrderApprovedEvent 发布存款订单审核通过事件到 Kafka。
// 对应 operator-activity 充值活动插件消费的 topic。
func (h *EventHandler) PublishDepositOrderApprovedEvent(ctx context.Context, in DepositApprovedEventInput) {
	if h.producer == nil {
		h.Errorf("[活动事件] Kafka 生产者未配置，跳过发送存款审核通过事件: member_id=%d order_id=%d", in.MemberID, in.OrderID)
		return
	}

	now := time.Now()
	data := depositPayload{
		OperatorCode:   in.OperatorCode,
		MemberID:       in.MemberID,
		OrderID:        in.OrderID,
		Amount:         in.Amount,
		ClientType:     in.ClientType,
		Currency:       in.Currency,
		Timestamp:      now.UnixMilli(),
		DepositCount:   in.DepositCount,
		EventType:      string(topic.GameUserDepositOrderApproved),
		OccurredAt:     now,
		PublishedAt:    now,
		MemberGlobalID: in.MemberGlobalID,
	}

	if err := h.producer.PublishData(ctx, topic.GameUserDepositOrderApproved, in.OperatorCode, data); err != nil {
		h.Errorf("[活动事件] 发送存款审核通过事件失败: member_id=%d order_id=%d err=%v", in.MemberID, in.OrderID, err)
		return
	}
	h.Infof("[活动事件] 成功发送存款审核通过事件: member_id=%d order_id=%d deposit_count=%d", in.MemberID, in.OrderID, in.DepositCount)
}

// PublishGameOrderSettledEvent 发布游戏订单结算事件到 Kafka。
// 对应 operator-activity 投注活动插件消费的 topic。
func (h *EventHandler) PublishGameOrderSettledEvent(ctx context.Context, in GameOrderSettledEventInput) {
	if h.producer == nil {
		h.Errorf("[活动事件] Kafka 生产者未配置，跳过发送游戏订单结算事件: member_id=%d order_id=%d", in.MemberID, in.OrderID)
		return
	}

	now := time.Now()
	data := betPayload{
		OperatorCode:   in.OperatorCode,
		MemberID:       in.MemberID,
		OrderID:        in.OrderID,
		Amount:         in.Amount,
		GameID:         in.GameID,
		ClientType:     in.ClientType,
		Currency:       in.Currency,
		Timestamp:      now.UnixMilli(),
		EventType:      string(topic.GameUserGameOrderSettled),
		OccurredAt:     now,
		PublishedAt:    now,
		MemberGlobalID: in.MemberGlobalID,
	}

	err := h.producer.PublishData(ctx, topic.GameUserGameOrderSettled, in.OperatorCode, data)
	if err != nil {
		h.Errorf("[活动事件] 发送游戏订单结算事件失败: member_id=%d order_id=%d err=%v", in.MemberID, in.OrderID, err)
		return
	}
	h.Infof("[活动事件] 成功发送游戏订单结算事件: member_id=%d order_id=%d amount=%d", in.MemberID, in.OrderID, in.Amount)
}
