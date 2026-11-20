// Package types 定义系统中的数据类型和消息结构。
package types

import (
	"encoding/json"
	"time"
)

type Event struct {
	Topic     string `json:"topic"`     // 事件所属的Kafka主题
	Partition int    `json:"partition"` // 事件所属的Kafka分区
	Offset    int64  `json:"offset"`    // 事件在Kafka分区中的偏移量
	KafkaID   string `json:"kafka_id"`  // Kafka 消息 ID
	// 事件基本信息
	// EventID      string    `json:"event_id"`      // 全局唯一事件ID
	EventType    string    `json:"event_type"`    // 事件类型：member.registered, payment.deposit_succeeded等
	OperatorCode string    `json:"operator_code"` // 分站标识，必填
	MemberID     int64     `json:"member_id"`     // 会员ID，可选
	OccurredAt   time.Time `json:"occurred_at"`   // 业务事实发生时间
	PublishedAt  time.Time `json:"published_at"`  // 事件发布时间
	PeriodKey    string    `json:"period_key"`    // 周期键（日/周/月/一次性）
	TraceID      string    `json:"trace_id"`      // 链路追踪ID
	// Payload      json.RawMessage `json:"payload"`       // 业务数据

	TaskCode   string `json:"task_code"`   // 任务编码
	OrderNo    string `json:"order_no"`    // 订单号
	Amount     int64  `json:"amount"`      // 金额
	RetryCount int    `json:"retry_count"` // 重试次数
	MaxRetry   int    `json:"max_retry"`   // 最大重试次数（默认6次）
}

// ========== CDC 事件 Payload ==========

// CDCEnvelope CDC 事件的通用信封结构（包含 before/after）
type CDCEnvelope struct {
	Before json.RawMessage `json:"before"` // 修改前数据（null 表示 insert）
	After  json.RawMessage `json:"after"`  // 修改后数据（null 表示 delete）
	Op     string          `json:"op"`     // 操作类型: c(create), u(update), d(delete)
	TSMs   int64           `json:"ts_ms"`  // 时间戳（毫秒）
}

type CommonPayload struct {
	ID           int64     `json:"id"`            // 数据 ID
	OperatorCode string    `json:"operator_code"` // 分站代码
	MemberID     int64     `json:"member_id"`     // 会员 ID
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// MemberPayload 会员基本信息表的数据结构
type MemberPayload struct {
	ID           int       `json:"id"`            // 数据 ID
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
	OperatorCode string    `json:"operator_code"` // 分站代码
	LastLoginAt  time.Time `json:"last_login_at"` // 最后登录时间
}

// MemberOAuthIdentityPayload 会员 OAuth 身份信息表的数据结构
type MemberOAuthIdentityPayload struct {
	CommonPayload
	Provider string `json:"provider"` // OAuth 提供商（如 google, facebook）
}

// MemberAuthPayload 会员认证信息表的数据结构（电话、邮箱）
type MemberAuthPayload struct {
	CommonPayload
	PhoneCiphertext string `json:"phone_ciphertext"` // 加密的电话号码
	EmailCiphertext string `json:"email_ciphertext"` // 加密的邮箱地址
}

// MemberProfilePayload 会员个人资料表的数据结构
type MemberProfilePayload struct {
	CommonPayload
	Nickname                 string `json:"nickname"`                   // 昵称
	RealNameCiphertext       string `json:"real_name_ciphertext"`       // 加密的真实姓名
	IdentityNumberCiphertext string `json:"identity_number_ciphertext"` // 加密的身份证号
	Birthday                 int    `json:"birthday"`                   // 生日（Unix 天数）
	AvatarID                 int    `json:"avatar_id"`                  // 头像 ID
}

// GameDepositPayload 游戏充值记录表的数据结构
type GameDepositPayload struct {
	CommonPayload
}

// GameWithdrawPayload 游戏提现记录表的数据结构
type GameWithdrawPayload struct {
	CommonPayload
}

// AgentApplyPayload 代理申请表的数据结构
type AgentApplyPayload struct {
	CommonPayload
	RegisterSource int `json:"register_source"` // 注册来源
	AuditStatus    int `json:"audit_status"`    // 审核状态
}

// GameChatLogPayload 游戏聊天日志表的数据结构
type GameChatLogPayload struct {
	CommonPayload
}

// CDCEventMetadata CDC 事件的元数据（从消息中解析）
type CDCEventMetadata struct {
	ID           int       // 数据 ID
	OperatorCode string    // 分站代码
	MemberID     int64     // 会员 ID
	UpdatedAt    time.Time // 更新时间
	Op           string    // 操作类型
	Table        string    // 表名
	TSMs         int64     // 时间戳
}
