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

// GetForUpdateTx 查询并锁定会员钱包
func (d *MemberWallet) GetForUpdateTx(ctx context.Context, tx *ent.Tx, memberGlobalID uuid.UUID) (*ent.MemberWallet, error) {
	return tx.MemberWallet.Query().
		Where(memberwallet.MemberGlobalIDEQ(memberGlobalID)). // 按会员全局标识查询
		ForUpdate().                                          // 锁定钱包记录, 避免并发资金操作
		Only(ctx)
}

// UpdateAvailableBalanceTx 更新会员可用余额
func (d *MemberWallet) UpdateAvailableBalanceTx(ctx context.Context, tx *ent.Tx, walletID, balance int64) error {
	_, err := tx.MemberWallet.UpdateOneID(walletID).
		SetBalance(balance). // 设置更新后的可用余额
		Save(ctx)
	return err
}
