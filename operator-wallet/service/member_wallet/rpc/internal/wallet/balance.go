package wallet

import "github.com/google/uuid"

// BalanceSnapshot 表示某一时刻的钱包余额快照
type BalanceSnapshot struct {
	WalletGlobalID uuid.UUID // 钱包全局唯一标识
	CurrencyCode   string    // 钱包结算币种

	Balance          int64 // 可用余额
	FrozenBalance    int64 // 冻结余额
	UnsettledBalance int64 // 游戏未结算余额

	Version int64 // 钱包状态版本
}

// balanceQueryMessage 表示一次钱包余额查询
type balanceQueryMessage struct {
	reply chan balanceQueryReply // 查询结果通知
}

// balanceQueryReply 表示余额查询结果
type balanceQueryReply struct {
	result BalanceSnapshot // 当前钱包余额快照
	err    error           // 查询错误
}

// newBalanceQueryMessage 创建余额查询消息
func newBalanceQueryMessage() balanceQueryMessage {
	return balanceQueryMessage{
		reply: make(chan balanceQueryReply, 1),
	}
}

// reject 将未执行的余额查询返回失败结果
func (m balanceQueryMessage) reject(err error) {
	if m.reply == nil {
		return
	}

	m.reply <- balanceQueryReply{err: err}
}

// newBalanceSnapshot 根据钱包运行时状态创建余额快照
func newBalanceSnapshot(state State) BalanceSnapshot {
	return BalanceSnapshot{
		WalletGlobalID:   state.WalletGlobalID,
		CurrencyCode:     state.CurrencyCode,
		Balance:          state.Balance,
		FrozenBalance:    state.FrozenBalance,
		UnsettledBalance: state.UnsettledBalance,
		Version:          state.Version,
	}
}
