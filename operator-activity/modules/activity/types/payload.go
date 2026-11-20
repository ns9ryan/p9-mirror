package types

import (
	"encoding/json"
	"fmt"
)

// Payload Kafka 通用消息信封。业务字段只放在 Data。
type Payload struct {
	ID        string          `json:"id"`            // 消息 ID
	Name      string          `json:"name"`          // 事件名称，建议用 topic 常量
	Data      json.RawMessage `json:"data"`          // 业务体 JSON
	Timestamp int64           `json:"timestamp"`     // Unix 毫秒
	V         int             `json:"v"`             // schema 版本
	Ext       json.RawMessage `json:"ext,omitempty"` // 扩展，可选
}

// UnmarshalData 将 Data 解到业务结构。
func (p Payload) UnmarshalData(dest any) error {
	if len(p.Data) == 0 || string(p.Data) == "null" {
		return fmt.Errorf("payload: empty data")
	}
	if err := json.Unmarshal(p.Data, dest); err != nil {
		return fmt.Errorf("payload: unmarshal data: %w", err)
	}
	return nil
}

// UnmarshalExt 将 Ext 解到目标结构。
func (p Payload) UnmarshalExt(dest any) error {
	if len(p.Ext) == 0 || string(p.Ext) == "null" {
		return fmt.Errorf("payload: empty ext")
	}
	if err := json.Unmarshal(p.Ext, dest); err != nil {
		return fmt.Errorf("payload: unmarshal ext: %w", err)
	}
	return nil
}

// RegisterPayload Kafka 注册事件消息
type RegisterPayload struct {
	OperatorCode string `json:"operator_code"`
	MemberID     int64  `json:"member_id"`
	InviterID    int64  `json:"inviter_id"`
	PromoRefID   int64  `json:"promo_ref_id"`
	Timestamp    int64  `json:"timestamp"`
}

// DepositPayload Kafka 充值审核通过消息
type DepositPayload struct {
	OperatorCode string `json:"operator_code"`
	MemberID     int64  `json:"member_id"`
	OrderID      int64  `json:"order_id"`
	Amount       int64  `json:"amount"`
	ClientType   int    `json:"client_type"`
	Currency     string `json:"currency"`
	Timestamp    int64  `json:"timestamp"`
	DepositCount int    `json:"deposit_count"`
}

// BetPayload Kafka 投注事件消息
type BetPayload struct {
	OperatorCode string `json:"operator_code"`
	MemberID     int64  `json:"member_id"`
	OrderID      int64  `json:"order_id"`
	Amount       int64  `json:"amount"`
	GameID       int64  `json:"game_id"`
	ClientType   int    `json:"client_type"`
	Timestamp    int64  `json:"timestamp"`
}
