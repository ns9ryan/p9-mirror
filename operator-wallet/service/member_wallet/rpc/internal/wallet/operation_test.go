package wallet

import (
	"testing"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/enum"
)

// TestNewOperationMessageCopiesChanges 验证消息不会共享调用方的余额变动切片
func TestNewOperationMessageCopiesChanges(t *testing.T) {
	params := RecordParams{
		OperationID:    uuid.New(),
		BusinessType:   enum.BusinessTypeDeposit,
		BusinessAction: enum.BusinessActionDepositCredit,
		BusinessNo:     "D10001",
		Changes: []BalanceChange{
			{
				BalanceType:  enum.BalanceTypeAvailable,
				ChangeAmount: 100,
			},
		},
	}

	message := newOperationMessage(params, 2)

	if message.partition != 2 {
		t.Fatalf("expected partition 2, got %d", message.partition)
	}

	// 修改调用方的切片, 不应影响已经创建的消息
	params.Changes[0].ChangeAmount = 200

	if got := message.params.Changes[0].ChangeAmount; got != 100 {
		t.Fatalf("message changes were modified: got %d, want 100", got)
	}

	// 修改消息内部的切片, 也不应反向影响调用方
	message.params.Changes[0].ChangeAmount = 300

	if got := params.Changes[0].ChangeAmount; got != 200 {
		t.Fatalf("original changes were modified: got %d, want 200", got)
	}
}

// TestNewOperationMessageBufferedReply 验证调用方不再等待时仍可发送结果
func TestNewOperationMessageBufferedReply(t *testing.T) {
	operationID := uuid.New()
	walletGlobalID := uuid.New()

	message := newOperationMessage(RecordParams{OperationID: operationID}, 0)

	if cap(message.reply) != 1 {
		t.Fatalf("expected reply capacity 1, got %d", cap(message.reply))
	}

	expected := OperationResult{
		OperationID:    operationID,
		WalletGlobalID: walletGlobalID,
		Version:        10,
		Balance:        900,
	}

	// 没有接收者时, Actor 也应能够立即发送一次处理结果
	select {
	case message.reply <- operationReply{result: expected}:
	default:
		t.Fatal("reply channel should accept one result without a receiver")
	}

	reply := <-message.reply

	if reply.err != nil {
		t.Fatalf("unexpected reply error: %v", reply.err)
	}

	if reply.result != expected {
		t.Fatalf("unexpected result: got %+v, want %+v", reply.result, expected)
	}
}
