package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/memberwalletledger"
)

// MemberWalletLedger 会员资金流水数据访问
type MemberWalletLedger struct {
	db *ent.Client // Ent数据库客户端
}

// CreateLedgerParams 创建资金流水参数
type CreateLedgerParams struct {
	WalletID int64 // 所属钱包ID

	// 余额变动
	BalanceType   int64 // 余额类型
	ChangeAmount  int64 // 变动金额
	BalanceBefore int64 // 变动前余额
	BalanceAfter  int64 // 变动后余额

	// 操作与业务来源
	OperationID    uuid.UUID // 钱包操作唯一标识
	BusinessType   int64     // 一级业务类型
	BusinessAction int64     // 具体操作类型
	BusinessNo     string    // 原始业务单号
}

// GetByBusinessTx 根据业务类型和业务单号查询流水
func (d *MemberWalletLedger) GetByBusinessTx(ctx context.Context, tx *ent.Tx, businessType, businessAction int64, businessNo string) (*ent.MemberWalletLedger, error) {
	return tx.MemberWalletLedger.Query().
		Where(
			memberwalletledger.BusinessTypeEQ(businessType),     // 一级业务类型: 1存款, 2提现, 3游戏, 4活动, 5任务, 6佣金, 7系统奖励
			memberwalletledger.BusinessActionEQ(businessAction), // 具体资金操作类型
			memberwalletledger.BusinessNoEQ(businessNo),         // 原始业务单号
		).
		Only(ctx)
}

// GetByOperationTx 根据钱包操作标识查询流水
func (d *MemberWalletLedger) GetByOperationTx(ctx context.Context, tx *ent.Tx, walletID int64, operationID uuid.UUID, balanceType int64) (*ent.MemberWalletLedger, error) {
	return tx.MemberWalletLedger.Query().
		Where(
			memberwalletledger.WalletIDEQ(walletID),       // 所属钱包ID
			memberwalletledger.OperationIDEQ(operationID), // 钱包资金操作唯一标识
			memberwalletledger.BalanceTypeEQ(balanceType), // 一级业务类型: 1存款, 2提现, 3游戏, 4活动, 5任务, 6佣金, 7系统奖励
		).
		Only(ctx)
}

// CreateTx 在当前事务中创建资金流水
// operator_code 由Ent Hook从可信上下文自动写入
func (d *MemberWalletLedger) CreateTx(ctx context.Context, tx *ent.Tx, in CreateLedgerParams) (*ent.MemberWalletLedger, error) {
	return tx.MemberWalletLedger.Create().
		SetWalletID(in.WalletID).             // 所属钱包ID
		SetBalanceType(in.BalanceType).       // 余额类型
		SetChangeAmount(in.ChangeAmount).     // 变动金额
		SetBalanceBefore(in.BalanceBefore).   // 变动前余额
		SetBalanceAfter(in.BalanceAfter).     // 变动后余额
		SetOperationID(in.OperationID).       // 钱包操作唯一标识
		SetBusinessType(in.BusinessType).     // 一级业务类型
		SetBusinessAction(in.BusinessAction). // 具体操作类型
		SetBusinessNo(in.BusinessNo).         // 原始业务单号
		Save(ctx)
}

// IsDepositCreditUniqueConflict 判断是否触发存款入账唯一约束
func IsDepositCreditUniqueConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}

	switch pgErr.ConstraintName {
	case "uk_member_wallet_ledger_deposit_credit",
		"uk_member_wallet_ledger_operation":
		return true
	default:
		return false
	}
}
