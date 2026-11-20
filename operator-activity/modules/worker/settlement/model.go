package settlement

import (
	"time"
)

// SettlementWorkflowParam 工作流参数
type SettlementWorkflowParam struct {
	CollectionNo      string    `json:"collection_no"`
	MerchantID        string    `json:"merchant_id"`
	TotalAmount       float64   `json:"total_amount"`
	ReceiptDate       time.Time `json:"receipt_date"`
	PartialSettlement bool      `json:"partial_settlement"`
	CurrentPhase      int32     `json:"current_phase"`   // 当前结算阶段
	AlreadySettled    float64   `json:"already_settled"` // 已结算金额
	WorkflowID        string    `json:"workflow_id"`
	ReceiptID         string    `json:"receipt_id"`
}

// SettlementWorkflowRes 增强的工作流结果
type SettlementWorkflowRes struct {
	CollectionNo    string                    `json:"collection_no"`
	SettledAmount   float64                   `json:"settled_amount"`
	PendingAmount   float64                   `json:"pending_amount"`
	TotalPhases     int32                     `json:"total_phases"`
	CompletedPhases int32                     `json:"completed_phases"`
	SettlementTime  time.Time                 `json:"settlement_time"`
	Records         []SettlementRecordSummary `json:"records"`
	ErrorMessage    string                    `json:"error_message,omitempty"`
}

// SettlementRecordSummary 结算记录摘要
type SettlementRecordSummary struct {
	Sequence int       `json:"sequence"`
	Amount   float64   `json:"amount"`
	Date     time.Time `json:"date"`
}
