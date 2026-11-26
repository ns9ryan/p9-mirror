package operation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/managepb"
)

// ConcurrentDepositConfig 同钱包并发存款测试配置
type ConcurrentDepositConfig struct {
	MemberGlobalID string        // 测试会员全局标识
	CurrencyCode   string        // 钱包币种
	Amount         int64         // 单笔入账金额, 币种最小单位
	Concurrency    int           // 并发请求数量
	Timeout        time.Duration // 单次 RPC 超时时间
}

// ConcurrentDepositResult 同钱包并发存款测试结果
type ConcurrentDepositResult struct {
	TotalRequests   int // 总请求数
	SuccessRequests int // 成功请求数
	FailedRequests  int // 失败请求数

	BeforeBalance int64 // 操作前可用余额
	AfterBalance  int64 // 操作后可用余额

	BeforeVersion int64 // 操作前钱包版本
	AfterVersion  int64 // 操作后钱包版本

	BalanceMatched bool // 最终余额是否符合预期
	VersionMatched bool // 最终版本是否符合预期

	Elapsed    time.Duration // 整体测试耗时
	AvgLatency time.Duration // 平均 RPC 耗时
	P95Latency time.Duration // RPC P95 耗时
	MaxLatency time.Duration // 最大 RPC 耗时

	Errors []string // 请求失败信息
}

// ConcurrentDeposit 对同一个钱包发起并发存款请求
func ConcurrentDeposit(ctx context.Context, client memberwalletrpc.MemberWalletClient, cfg ConcurrentDepositConfig) (ConcurrentDepositResult, error) {
	if ctx == nil {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: nil context")
	}
	if client == nil {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: nil RPC client")
	}
	// 测试会员全局标识
	if cfg.MemberGlobalID == "" {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: member global ID is required")
	}
	// 钱包币种
	if cfg.CurrencyCode == "" {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: currency code is required")
	}
	// 单笔入账金额, 币种最小单位
	if cfg.Amount <= 0 {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: amount must be greater than zero")
	}
	// 并发请求数量
	if cfg.Concurrency <= 0 {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: concurrency must be greater than zero")
	}
	// 单次 RPC 超时时间
	if cfg.Timeout <= 0 {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: timeout must be greater than zero")
	}

	// 查询操作前余额
	beforeCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	before, err := client.GetBalance(beforeCtx, &managepb.GetBalanceRequest{
		MemberGlobalId: cfg.MemberGlobalID, // 会员全局唯一标识
	})
	cancel()
	if err != nil {
		return ConcurrentDepositResult{}, fmt.Errorf("get balance before concurrent deposit: %w", err)
	}

	// 校验钱包币种是否一致
	if before.GetCurrencyCode() != cfg.CurrencyCode {
		return ConcurrentDepositResult{}, fmt.Errorf("wallet currency mismatch: expected=%s actual=%s",
			cfg.CurrencyCode,
			before.GetCurrencyCode(),
		)
	}

	// 预先检查最终余额和版本是否可能溢出
	if int64(cfg.Concurrency) > math.MaxInt64/cfg.Amount {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: total amount overflow")
	}
	totalAmount := int64(cfg.Concurrency) * cfg.Amount

	if before.GetBalance() > math.MaxInt64-totalAmount {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: expected balance overflow")
	}
	if int64(cfg.Concurrency) > math.MaxInt64-before.GetVersion() {
		return ConcurrentDepositResult{}, errors.New("concurrent deposit: expected version overflow")
	}

	// 单次 RPC 请求结果
	type requestResult struct {
		err     error         // 单次请求错误
		elapsed time.Duration // 单次请求耗时
	}

	results := make([]requestResult, cfg.Concurrency)

	// 使用启动信号, 尽量让所有请求同时发起
	startCh := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(cfg.Concurrency)

	startedAt := time.Now()

	for i := 0; i < cfg.Concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			<-startCh

			operationID := uuid.NewString()
			orderNo := "LOADTEST-" + uuid.NewString()

			rpcCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
			defer cancel()

			requestStartedAt := time.Now()

			_, err := client.DepositCredit(rpcCtx, &depositpb.DepositCreditRequest{
				MemberGlobalId: cfg.MemberGlobalID, // 会员全局唯一标识
				OperationId:    operationID,        // 独立操作标识
				DepositOrderNo: orderNo,            // 独立存款订单号
				CurrencyCode:   cfg.CurrencyCode,   // 订单币种
				Amount:         cfg.Amount,         // 入账金额
			})

			results[index] = requestResult{
				err:     err,
				elapsed: time.Since(requestStartedAt),
			}
		}(i)
	}

	// 释放所有 Goroutine, 开始并发请求
	close(startCh)
	wg.Wait()

	elapsed := time.Since(startedAt)

	// 查询操作后余额
	afterCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	after, err := client.GetBalance(afterCtx, &managepb.GetBalanceRequest{
		MemberGlobalId: cfg.MemberGlobalID, // 会员全局唯一标识
	})
	cancel()
	if err != nil {
		return ConcurrentDepositResult{}, fmt.Errorf("get balance after concurrent deposit: %w", err)
	}

	result := ConcurrentDepositResult{
		TotalRequests: cfg.Concurrency, // 总请求数

		BeforeBalance: before.GetBalance(), // 操作前可用余额
		AfterBalance:  after.GetBalance(),  // 操作后可用余额

		BeforeVersion: before.GetVersion(), // 操作前钱包版本
		AfterVersion:  after.GetVersion(),  // 操作后钱包版本

		Elapsed: elapsed, // 整体测试耗时
	}

	// 汇总请求结果
	latencies := make([]time.Duration, 0, cfg.Concurrency)
	var totalLatency time.Duration

	for i, item := range results {
		if item.err != nil {
			result.FailedRequests++ // 失败请求数
			result.Errors = append(result.Errors, fmt.Sprintf("request %d: %v", i+1, item.err))
			continue
		}

		result.SuccessRequests++ // 成功请求数
		latencies = append(latencies, item.elapsed)
		totalLatency += item.elapsed
	}

	// 统计成功请求的 RPC 耗时
	if len(latencies) > 0 {
		slices.Sort(latencies)

		result.AvgLatency = totalLatency / time.Duration(len(latencies)) // 平均 RPC 耗时
		result.P95Latency = latencies[(len(latencies)*95+99)/100-1]      // RPC P95 耗时
		result.MaxLatency = latencies[len(latencies)-1]                  // 最大 RPC 耗时
	}

	// 只有全部请求成功时, 才能直接按请求数核对最终余额和版本
	if result.FailedRequests == 0 {
		expectedBalance := before.GetBalance() + totalAmount
		expectedVersion := before.GetVersion() + int64(cfg.Concurrency)

		result.BalanceMatched = after.GetBalance() == expectedBalance // 最终余额是否符合预期
		result.VersionMatched = after.GetVersion() == expectedVersion // 最终版本是否符合预期
	}

	return result, nil
}
