package xkafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DefaultPayloadVersion 默认的 payload 版本
const DefaultPayloadVersion = 1

// Payload Kafka 通用消息信封。业务字段只放在 Data。
type Payload struct {
	ID        string          `json:"id"`            // 消息 ID，NewPayload 自动生成
	Name      string          `json:"name"`          // 事件名称，建议用 topic 常量
	Data      json.RawMessage `json:"data"`          // 业务体 JSON
	Timestamp int64           `json:"time"`          // Unix 毫秒
	V         int             `json:"v"`             // schema 版本
	Ext       json.RawMessage `json:"ext,omitempty"` // 扩展，可选
}

// NewPayload 构造信封：ID 自动生成，Time=当前毫秒，V=1。
func NewPayload(name string, data any) (Payload, error) {
	return NewPayloadExt(name, data, nil)
}

// NewPayloadExt 构造带扩展字段的信封。
func NewPayloadExt(name string, data, ext any) (Payload, error) {
	raw, err := marshalField(data)
	if err != nil {
		return Payload{}, fmt.Errorf("xkafka: marshal data: %w", err)
	}
	extRaw, err := marshalField(ext)
	if err != nil {
		return Payload{}, fmt.Errorf("xkafka: marshal ext: %w", err)
	}
	return Payload{
		ID:        newMessageID(),
		Name:      name,
		Data:      raw,
		Timestamp: time.Now().UnixMilli(),
		V:         DefaultPayloadVersion,
		Ext:       extRaw,
	}, nil
}

// NewPayload 构造信封：ID 自动生成，Time=当前毫秒，V=1。
func NewPayloadWithID(id string, name string, data any) (Payload, error) {
	return NewPayloadExtWithID(id, name, data, nil)
}

// NewPayloadExt 构造带扩展字段的信封。
func NewPayloadExtWithID(id string, name string, data, ext any) (Payload, error) {
	raw, err := marshalField(data)
	if err != nil {
		return Payload{}, fmt.Errorf("xkafka: marshal data: %w", err)
	}
	extRaw, err := marshalField(ext)
	if err != nil {
		return Payload{}, fmt.Errorf("xkafka: marshal ext: %w", err)
	}
	return Payload{
		ID:        id,
		Name:      name,
		Data:      raw,
		Timestamp: time.Now().UnixMilli(),
		V:         DefaultPayloadVersion,
		Ext:       extRaw,
	}, nil
}

// newMessageID 生成 UUID v4 消息 ID。
func newMessageID() string {
	return uuid.NewString()
}

// DecodePayload 从消息 value 解析信封。
func DecodePayload(value []byte) (Payload, error) {
	if len(value) == 0 {
		return Payload{}, fmt.Errorf("xkafka: empty payload")
	}
	var p Payload
	if err := json.Unmarshal(value, &p); err != nil {
		return Payload{}, fmt.Errorf("xkafka: decode payload: %w", err)
	}
	return p, nil
}

// UnmarshalData 将 Data 解到业务结构。
func (p Payload) UnmarshalData(dest any) error {
	if len(p.Data) == 0 || string(p.Data) == "null" {
		return fmt.Errorf("xkafka: empty data")
	}
	if err := json.Unmarshal(p.Data, dest); err != nil {
		return fmt.Errorf("xkafka: unmarshal data: %w", err)
	}
	return nil
}

// UnmarshalExt 将 Ext 解到目标结构。
func (p Payload) UnmarshalExt(dest any) error {
	if len(p.Ext) == 0 || string(p.Ext) == "null" {
		return fmt.Errorf("xkafka: empty ext")
	}
	if err := json.Unmarshal(p.Ext, dest); err != nil {
		return fmt.Errorf("xkafka: unmarshal ext: %w", err)
	}
	return nil
}

// marshalField 将任意类型序列化为 JSON 字节切片。
func marshalField(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	if raw, ok := v.(json.RawMessage); ok {
		return raw, nil
	}
	if raw, ok := v.([]byte); ok {
		if !json.Valid(raw) {
			return nil, fmt.Errorf("invalid json")
		}
		return json.RawMessage(raw), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}
