package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/memberwallet"
)

// MemberWallet 会员钱包数据访问
type MemberWallet struct {
	db *ent.Client // Ent数据库客户端
}

var (
	// ErrWalletIdentityConflict 钱包已存在, 但会员标识或币种不一致
	ErrWalletIdentityConflict = errors.New("wallet identity conflict")
)

// CreateWalletParams 创建会员钱包参数
type CreateWalletParams struct {
	MemberID       int64     // CRM会员ID
	MemberGlobalID uuid.UUID // CRM会员全局唯一标识
	CurrencyCode   string    // 钱包结算币种编码
}

// CreateOrGet 创建会员钱包, 已存在且信息一致时返回原钱包
func (d *MemberWallet) CreateOrGet(ctx context.Context, in CreateWalletParams) (*ent.MemberWallet, error) {
	// 首次创建使用 Ent 默认值初始化余额、版本和钱包全局标识
	// operator_code 由 Ent Hook 从可信 Context 自动绑定
	created, err := d.db.MemberWallet.Create().
		SetMemberID(in.MemberID).
		SetMemberGlobalID(in.MemberGlobalID).
		SetCurrencyCode(in.CurrencyCode).
		Save(ctx)
	if err == nil {
		return created, nil
	}

	// 只有唯一约束冲突才需要查询已有钱包
	if !ent.IsConstraintError(err) {
		return nil, fmt.Errorf("create member wallet: %w", err)
	}

	// 同时检查两个会员标识, 避免不同标识命中不同钱包
	existing, lookupErr := d.db.MemberWallet.Query().
		Where(
			memberwallet.Or(
				memberwallet.MemberIDEQ(in.MemberID),
				memberwallet.MemberGlobalIDEQ(in.MemberGlobalID),
			),
		).
		All(ctx)
	if lookupErr != nil {
		return nil, fmt.Errorf("load existing member wallet: %w", lookupErr)
	}

	// 当前厅没有匹配记录时, 仍视为唯一标识冲突
	if len(existing) != 1 {
		return nil, ErrWalletIdentityConflict
	}

	current := existing[0]

	// 只有会员标识和币种全部一致才视为幂等成功
	if current.MemberID != in.MemberID || current.MemberGlobalID != in.MemberGlobalID || current.CurrencyCode != in.CurrencyCode {
		return nil, ErrWalletIdentityConflict
	}

	return current, nil
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
