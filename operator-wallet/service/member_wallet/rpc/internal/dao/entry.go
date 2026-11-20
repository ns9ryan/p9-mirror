package dao

import "oa.98ent.com/p9/operator-wallet/ent"

// Group 统一管理钱包数据访问
type Group struct {
	MemberWallet       MemberWallet       // 会员钱包
	MemberWalletLedger MemberWalletLedger // 会员资金流水
}

// New 创建数据访问入口
func New(db *ent.Client) *Group {
	return &Group{
		MemberWallet:       MemberWallet{db: db},
		MemberWalletLedger: MemberWalletLedger{db: db},
	}
}
