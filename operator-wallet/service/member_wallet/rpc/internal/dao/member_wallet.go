package dao

import (
	"context"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/memberwallet"
)

// MemberWallet 会员钱包数据访问
type MemberWallet struct {
	db *ent.Client // Ent数据库客户端
}

// UpdateBalancesParams 更新钱包余额和版本参数
type UpdateBalancesParams struct {
	WalletID         int64 // 钱包ID
	Balance          int64 // 更新后的可用余额
	FrozenBalance    int64 // 更新后的冻结余额
	UnsettledBalance int64 // 更新后的游戏未结算余额
	Version          int64 // 更新后的钱包状态版本
}

// GetByMemberGlobalID 根据会员全局标识查询当前厅的钱包
func (d *MemberWallet) GetByMemberGlobalID(ctx context.Context, memberGlobalID uuid.UUID) (*ent.MemberWallet, error) {
	return d.db.MemberWallet.Query().
		Where(memberwallet.MemberGlobalIDEQ(memberGlobalID)). // CRM会员全局唯一标识
		Only(ctx)
}

// GetForUpdateTx 查询并锁定会员钱包
func (d *MemberWallet) GetForUpdateTx(ctx context.Context, tx *ent.Tx, memberGlobalID uuid.UUID) (*ent.MemberWallet, error) {
	return tx.MemberWallet.Query().
		Where(memberwallet.MemberGlobalIDEQ(memberGlobalID)). // 按会员全局标识查询
		ForUpdate().                                          // 锁定钱包记录, 避免并发资金操作
		Only(ctx)
}

// UpdateBalancesTx 在当前事务中一次更新钱包余额和版本
func (d *MemberWallet) UpdateBalancesTx(ctx context.Context, tx *ent.Tx, in UpdateBalancesParams) error {
	_, err := tx.MemberWallet.UpdateOneID(in.WalletID).
		SetBalance(in.Balance).                   // 更新后的可用余额
		SetFrozenBalance(in.FrozenBalance).       // 更新后的冻结余额
		SetUnsettledBalance(in.UnsettledBalance). // 更新后的未结算余额
		SetVersion(in.Version).                   // 更新后的钱包状态版本
		Save(ctx)
	return err
}
