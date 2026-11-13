package wallet

import (
	"slices"

	"github.com/google/uuid"
)

// operationMessage 表示提交给 Actor 的一次资金操作
type operationMessage struct {
	params    RecordParams        // 资金操作参数
	partition int32               // 目标 Kafka Partition
	reply     chan operationReply // 资金操作结果通知
}

// operationReply 表示 Actor 返回的资金操作处理结果
type operationReply struct {
	result OperationResult // 资金操作成功后的结果
	err    error           // 资金操作失败或结果不明确时的错误
}

// OperationResult 表示一次已确认资金操作的结果
type OperationResult struct {
	// 操作与钱包标识
	OperationID    uuid.UUID // 资金操作唯一标识
	WalletGlobalID uuid.UUID // 钱包全局唯一标识

	// 操作完成时的钱包状态
	Version          int64 // 本次操作完成后的钱包版本
	Balance          int64 // 可用余额
	FrozenBalance    int64 // 冻结余额
	UnsettledBalance int64 // 游戏未结算余额

	Replayed bool // 是否返回此前已成功操作的幂等结果
}

// newOperationMessage 创建资金操作消息
func newOperationMessage(params RecordParams, partition int32) operationMessage {
	// 复制余额变动, 避免调用方修改已创建的消息
	params.Changes = slices.Clone(params.Changes)

	return operationMessage{
		params:    params,
		partition: partition,
		reply:     make(chan operationReply, 1),
	}
}
