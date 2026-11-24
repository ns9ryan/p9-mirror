package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
)

var (
	ErrWalletPersistenceUnavailable = errors.New("wallet persistence unavailable") // 持久化组件不可用
	ErrWalletIdentityMismatch       = errors.New("wallet identity mismatch")       // 钱包身份不一致
	ErrWalletVersionMismatch        = errors.New("wallet version mismatch")        // 钱包版本不一致
	ErrWalletBalanceMismatch        = errors.New("wallet balance mismatch")        // 钱包余额不一致
	ErrWalletOffsetMismatch         = errors.New("wallet offset mismatch")         // 消费进度不一致
)

// Persistence 负责 Kafka 资金事件的数据库事务持久化
type Persistence struct {
	db         *ent.Client                 // Ent数据库客户端
	wallet     *dao.MemberWallet           // 钱包数据访问
	ledger     *dao.MemberWalletLedger     // 资金流水数据访问
	checkpoint *dao.MemberWalletCheckpoint // Kafka消费进度数据访问

	topic  string  // 资金事件Topic
	router *router // 钱包固定分区路由
}

// NewPersistence 创建钱包资金事件持久化组件
func NewPersistence(db *ent.Client, topic string, shardCount int) (*Persistence, error) {
	if db == nil {
		return nil, ErrWalletPersistenceUnavailable
	}

	if topic == "" || strings.TrimSpace(topic) != topic {
		return nil, fmt.Errorf("%w: invalid topic", ErrWalletPersistenceUnavailable)
	}

	r, err := newRouter(shardCount)
	if err != nil {
		return nil, err
	}

	walletDAO := dao.New(db)

	return &Persistence{
		db:         db,
		wallet:     &walletDAO.MemberWallet,
		ledger:     &walletDAO.MemberWalletLedger,
		checkpoint: dao.NewMemberWalletCheckpoint(db),

		topic:  topic,
		router: r,
	}, nil
}

// Handle 校验并持久化单条 Kafka 资金事件
func (p *Persistence) Handle(ctx context.Context, record *kgo.Record) error {
	if p == nil || p.db == nil {
		return ErrWalletPersistenceUnavailable
	}
	if ctx == nil {
		return fmt.Errorf("%w: nil context", ErrWalletPersistenceUnavailable)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	// 校验消息所属 Topic
	if record == nil || record.Topic != p.topic {
		return invalidWalletEvent("unexpected Kafka topic")
	}

	// 解码并校验资金事件完整性
	event, err := decodeWalletEvent(record, p.router)
	if err != nil {
		return err
	}

	// 根据内部资金事件设置厅上下文, 事务内仍需校验数据库钱包归属
	ctx = ctxdata.WithOperatorCode(ctx, event.OperatorCode)

	// 钱包余额、资金流水及消费进度必须在同一事务中提交
	return dao.WithTx(ctx, p.db, func(tx *ent.Tx) error {
		return p.persistTx(ctx, tx, record, event)
	})
}

// persistTx 在当前事务中持久化完整资金事件
func (p *Persistence) persistTx(ctx context.Context, tx *ent.Tx, record *kgo.Record, event walletEvent) error {
	// 查询并锁定当前 Partition 的消费进度
	checkpoint, err := p.checkpoint.GetForUpdateTx(ctx, tx, record.Topic, record.Partition)
	if err != nil {
		return fmt.Errorf("lock wallet checkpoint: %w", err)
	}

	// 已经持久化的消息不重复处理
	if record.Offset < checkpoint.NextOffset {
		return nil
	}

	// 当前阶段禁止未经确认的 Offset 跳跃
	if record.Offset > checkpoint.NextOffset {
		return fmt.Errorf(
			"%w: partition=%d expected=%d actual=%d",
			ErrWalletOffsetMismatch, record.Partition, checkpoint.NextOffset, record.Offset,
		)
	}

	// 查询并锁定会员钱包
	memberWallet, err := p.wallet.GetForUpdateTx(ctx, tx, event.MemberGlobalID)
	if err != nil {
		return fmt.Errorf("lock member wallet: %w", err)
	}

	// 校验数据库钱包与事件操作前状态是否一致
	if err := validateWalletBeforeEvent(memberWallet, event); err != nil {
		return err
	}

	// 更新钱包余额及版本, 直接使用 Kafka 事件中的操作后快照
	if err := p.wallet.UpdateBalancesTx(ctx, tx, dao.UpdateBalancesParams{
		WalletID:         memberWallet.ID,              // 数据库钱包ID
		Balance:          event.After.Balance,          // 操作后可用余额
		FrozenBalance:    event.After.FrozenBalance,    // 操作后冻结余额
		UnsettledBalance: event.After.UnsettledBalance, // 操作后未结算余额
		Version:          event.Version,                // 操作后钱包版本
	}); err != nil {
		return fmt.Errorf("update wallet balances: %w", err)
	}

	// 保存本次操作产生的全部资金流水
	ledgers := buildLedgerParams(memberWallet, event)
	if err := p.ledger.CreateManyTx(ctx, tx, ledgers); err != nil {
		return fmt.Errorf("create wallet ledgers: %w", err)
	}

	// 资金数据全部写入成功后推进消费进度
	if err := p.checkpoint.AdvanceTx(
		ctx,
		tx,
		record.Topic,
		record.Partition,
		checkpoint.NextOffset,
		record.Offset,
	); err != nil {
		return fmt.Errorf("advance wallet checkpoint: %w", err)
	}

	return nil
}

// validateWalletBeforeEvent 校验数据库钱包身份、版本及操作前余额
func validateWalletBeforeEvent(memberWallet *ent.MemberWallet, event walletEvent) error {
	if memberWallet == nil {
		return ErrWalletIdentityMismatch
	}

	// 校验钱包及会员归属, 防止资金事件写入错误钱包
	if memberWallet.ID != event.WalletID ||
		memberWallet.GlobalID != event.WalletGlobalID ||
		memberWallet.MemberGlobalID != event.MemberGlobalID ||
		memberWallet.OperatorCode != event.OperatorCode ||
		memberWallet.CurrencyCode != event.CurrencyCode {
		return fmt.Errorf("%w: wallet_global_id=%s", ErrWalletIdentityMismatch, event.WalletGlobalID)
	}

	// 数据库版本必须与本次操作前版本一致
	if memberWallet.Version != event.PreviousVersion {
		return fmt.Errorf(
			"%w: wallet_global_id=%s expected=%d actual=%d",
			ErrWalletVersionMismatch, event.WalletGlobalID, event.PreviousVersion, memberWallet.Version,
		)
	}

	// 数据库三类余额必须与操作前快照完全一致
	if memberWallet.Balance != event.Before.Balance ||
		memberWallet.FrozenBalance != event.Before.FrozenBalance ||
		memberWallet.UnsettledBalance != event.Before.UnsettledBalance {
		return fmt.Errorf("%w: wallet_global_id=%s version=%d", ErrWalletBalanceMismatch, event.WalletGlobalID, event.PreviousVersion)
	}

	return nil
}

// buildLedgerParams 根据原始资金事件构造流水持久化参数
func buildLedgerParams(memberWallet *ent.MemberWallet, event walletEvent) []dao.CreateLedgerParams {
	items := make([]dao.CreateLedgerParams, 0, len(event.Ledgers))

	for _, ledger := range event.Ledgers {
		items = append(items, dao.CreateLedgerParams{
			// 流水及钱包归属
			GlobalID: ledger.LedgerGlobalID, // 原始流水全局唯一标识
			WalletID: memberWallet.ID,       // 数据库钱包ID
			MemberID: memberWallet.MemberID, // 数据库CRM会员ID

			// 余额变动
			BalanceType:   int64(ledger.BalanceType), // 余额类型
			ChangeAmount:  ledger.ChangeAmount,       // 变动金额
			BalanceBefore: ledger.BalanceBefore,      // 变动前余额
			BalanceAfter:  ledger.BalanceAfter,       // 变动后余额

			// 操作与业务来源
			OperationID:    event.OperationID,           // 钱包操作唯一标识
			BusinessType:   int64(event.BusinessType),   // 一级业务类型
			BusinessAction: int64(event.BusinessAction), // 具体资金操作类型
			BusinessNo:     event.BusinessNo,            // 原始业务单号

			CreatedAt: event.CreatedAt, // 原始资金事件生成时间
		})
	}

	return items
}
