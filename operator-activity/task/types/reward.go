package types

import (
	"context"
	"time"
)

// WalletClient 钱包服务客户端接口 TODO 等钱包grpc接口实现后移除该接口
type WalletClient interface {
	// Payout 派发奖励到会员钱包
	Payout(ctx context.Context, req *PayoutRequest) (*PayoutResponse, error)

	// QueryPayoutStatus 查询派发状态
	QueryPayoutStatus(ctx context.Context, walletOrderNo string) (*PayoutStatus, error)
}

// PayoutRequest 派发请求
type PayoutRequest struct {
	// 基本信息
	OpCode   string
	MemberID int64
	AgentID  int64

	// 订单信息
	WalletOrderNo   string // 幂等单号
	BusinessType    string // TASK_REWARD / ACTIVITY_REWARD
	BusinessOrderID string // 业务订单ID

	// 金额信息
	Currency    string // 货币代码
	AmountMinor int64  // 最小单位金额

	// 追踪
	TraceID string
	Remark  string
}

// PayoutResponse 派发响应
type PayoutResponse struct {
	Success             bool
	Code                string
	Message             string
	WalletID            string
	WalletTransactionID string
	WalletOrderNo       string
	PayoutAt            time.Time
	ErrorCode           string
	ErrorMessage        string
	Retryable           bool // 是否可重试
}

// PayoutStatus 派发状态
type PayoutStatus struct {
	Status              string
	AmountMinor         int64
	Currency            string
	PayoutAt            time.Time
	UpdatedAt           time.Time
	WalletTransactionID string
	ErrorCode           string
}

// Payout派发结果回调（用于Temporal Activity）
type PayoutResult struct {
	OrderNo             string
	Success             bool
	WalletTransactionID string
	ErrorCode           string
	ErrorMessage        string
	RetryAfter          time.Duration // 建议重试等待时间
}

type RewardWorkflowInput struct {
	RewardOrderNo    string
	OperatorCode     string
	MemberID         int64
	Code             string
	OrderID          int64
	Amount           int64
	Currency         string
	EffectAt         *time.Time
	TurnoverRequired int64
}

type RewardWorkflowOutput struct {
	WalletTransactionID string
	RewardOrderNo       string
	Success             bool
	ErrorCode           string
	ErrorMessage        string
}

type PayoutActivityInput struct {
	RewardOrderNo    string
	OperatorCode     string
	MemberID         int64
	Code             string
	OrderID          int64
	Amount           int64
	Currency         string
	EffectAt         *time.Time
	TurnoverRequired int64
}

// ========== Activity Output 结构体 ==========

// PayoutActivityOutput 派发Activity的输出
type PayoutActivityOutput struct {
	Success             bool
	WalletTransactionID string
	ErrorCode           string
	ErrorMessage        string
}

// BatchPayoutActivityOutput 批量派发Activity的输出
type BatchPayoutActivityOutput struct {
	Success      bool
	Count        int // 成功派发的订单数
	ErrorCode    string
	ErrorMessage string
}

// RetryActivityOutput 重试Activity的输出
type RetryActivityOutput struct {
	Success      bool
	Count        int // 重试成功的订单数
	ErrorCode    string
	ErrorMessage string
}

type PayoutTask struct {
	OrderNo      string `json:"order_no"`       // 订单号
	OperatorCode string `json:"operator_code"`  // 分站编码
	MemberID     int64  `json:"member_id"`      // 会员 ID
	Amount       int64  `json:"amount"`         // 派发金额
	TaskCode     string `json:"task_code"`      // 任务编码
	RetryCount   int    `json:"retry_count"`    // 重试次数
	MaxRetry     int    `json:"max_retry"`      // 最大重试次数
	LastErrorMsg string `json:"last_error_msg"` // 上一次错误信息
}
