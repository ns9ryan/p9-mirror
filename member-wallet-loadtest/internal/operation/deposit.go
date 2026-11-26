package operation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/managepb"
)

// DepositConfig 单笔存款测试配置
type DepositConfig struct {
	MemberGlobalID string        // 测试会员全局标识
	CurrencyCode   string        // 钱包币种
	Amount         int64         // 入账金额, 币种最小单位
	Timeout        time.Duration // 单次 RPC 超时时间
}

// DepositResult 单笔存款测试结果
type DepositResult struct {
	OperationID string // 钱包操作唯一标识
	OrderNo     string // 存款订单号

	BeforeBalance int64 // 操作前可用余额
	AfterBalance  int64 // 操作后可用余额

	BeforeVersion int64 // 操作前钱包版本
	AfterVersion  int64 // 操作后钱包版本

	Elapsed time.Duration // 本次测试耗时
}

// Deposit 执行单笔存款并验证运行时余额
func Deposit(ctx context.Context, client memberwalletrpc.MemberWalletClient, cfg DepositConfig) (DepositResult, error) {
	if ctx == nil {
		return DepositResult{}, errors.New("deposit: nil context")
	}
	if client == nil {
		return DepositResult{}, errors.New("deposit: nil RPC client")
	}
	if cfg.MemberGlobalID == "" {
		return DepositResult{}, errors.New("deposit: member global ID is required")
	}
	if cfg.CurrencyCode == "" {
		return DepositResult{}, errors.New("deposit: currency code is required")
	}
	if cfg.Amount <= 0 {
		return DepositResult{}, errors.New("deposit: amount must be greater than zero")
	}
	if cfg.Timeout <= 0 {
		return DepositResult{}, errors.New("deposit: timeout must be greater than zero")
	}

	operationID := uuid.NewString()           // 钱包操作唯一标识
	orderNo := "LOADTEST-" + uuid.NewString() // 存款订单号

	startedAt := time.Now()

	// 查询操作前余额
	beforeCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	before, err := client.GetBalance(beforeCtx, &managepb.GetBalanceRequest{
		MemberGlobalId: cfg.MemberGlobalID, // 会员全局唯一标识
	})
	cancel()
	if err != nil {
		return DepositResult{}, fmt.Errorf("get balance before deposit: %w", err)
	}

	// 判断币种是否一致
	if before.GetCurrencyCode() != cfg.CurrencyCode {
		return DepositResult{}, fmt.Errorf(
			"wallet currency mismatch: expected=%s actual=%s",
			cfg.CurrencyCode,
			before.GetCurrencyCode(),
		)
	}

	// 判断余额是否超出最大值
	if before.GetBalance() > math.MaxInt64-cfg.Amount {
		return DepositResult{}, errors.New("deposit: expected balance overflow")
	}

	// 判断钱包状态版本是否超出最大值
	if before.GetVersion() == math.MaxInt64 {
		return DepositResult{}, errors.New("deposit: expected version overflow")
	}

	// 提交存款本金入账
	depositCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	_, err = client.DepositCredit(depositCtx, &depositpb.DepositCreditRequest{
		MemberGlobalId: cfg.MemberGlobalID, // 会员全局唯一标识
		OperationId:    operationID,        // 钱包操作唯一标识, UUID
		DepositOrderNo: orderNo,            // 存款订单号
		CurrencyCode:   cfg.CurrencyCode,   // 订单币种编码
		Amount:         cfg.Amount,         // 入账金额, 币种最小单位
	})
	cancel()
	if err != nil {
		return DepositResult{}, fmt.Errorf("deposit credit: %w", err)
	}

	// 查询操作后余额
	afterCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	after, err := client.GetBalance(afterCtx, &managepb.GetBalanceRequest{
		MemberGlobalId: cfg.MemberGlobalID, // 会员全局唯一标识
	})
	cancel()
	if err != nil {
		return DepositResult{}, fmt.Errorf("get balance after deposit: %w", err)
	}

	result := DepositResult{
		OperationID: operationID, // 钱包操作唯一标识
		OrderNo:     orderNo,     // 存款订单号

		BeforeBalance: before.GetBalance(), // 操作前可用余额
		AfterBalance:  after.GetBalance(),  // 操作后可用余额

		BeforeVersion: before.GetVersion(), // 操作前钱包版本
		AfterVersion:  after.GetVersion(),  // 操作后钱包版本

		Elapsed: time.Since(startedAt), // 本次测试耗时
	}

	// 验证本次操作的余额和版本变化
	expectedBalance := before.GetBalance() + cfg.Amount
	expectedVersion := before.GetVersion() + 1

	if after.GetBalance() != expectedBalance {
		return result, fmt.Errorf(
			"deposit balance mismatch: expected=%d actual=%d",
			expectedBalance,
			after.GetBalance(),
		)
	}

	if after.GetVersion() != expectedVersion {
		return result, fmt.Errorf(
			"deposit version mismatch: expected=%d actual=%d",
			expectedVersion,
			after.GetVersion(),
		)
	}

	return result, nil
}
