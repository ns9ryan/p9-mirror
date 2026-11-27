package operation

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/depositpb"

	"oa.98ent.com/p9/member-wallet-loadtest/internal/setup"
)

// MultiWalletDepositConfig 多钱包并发存款测试配置
type MultiWalletDepositConfig struct {
	Wallets     []setup.Wallet // 参与测试的钱包
	Amount      int64          // 单笔入账金额, 币种最小单位
	Concurrency int            // 最大并发请求数
	Timeout     time.Duration  // 单次RPC超时时间
}

// MultiWalletDepositResult 多钱包并发存款测试结果
type MultiWalletDepositResult struct {
	TotalRequests   int // 总请求数
	SuccessRequests int // 成功请求数
	FailedRequests  int // 失败请求数
	SkippedRequests int // 未执行请求数

	Elapsed time.Duration // 整体测试耗时
	TPS     float64       // 每秒成功请求数

	AvgLatency time.Duration // 平均请求耗时
	P50Latency time.Duration // P50请求耗时
	P95Latency time.Duration // P95请求耗时
	P99Latency time.Duration // P99请求耗时
	MaxLatency time.Duration // 最大请求耗时

	Errors []string // 失败请求信息
}

// MultiWalletDeposit 对不同会员钱包执行并发存款
func MultiWalletDeposit(ctx context.Context, client memberwalletrpc.MemberWalletClient, cfg MultiWalletDepositConfig) (MultiWalletDepositResult, error) {
	if ctx == nil {
		return MultiWalletDepositResult{}, errors.New("multi-wallet deposit: nil context")
	}
	if client == nil {
		return MultiWalletDepositResult{}, errors.New("multi-wallet deposit: nil RPC client")
	}
	if len(cfg.Wallets) == 0 {
		return MultiWalletDepositResult{}, errors.New("multi-wallet deposit: wallets are required")
	}
	if cfg.Amount <= 0 {
		return MultiWalletDepositResult{}, errors.New("multi-wallet deposit: amount must be greater than zero")
	}
	if cfg.Concurrency <= 0 {
		return MultiWalletDepositResult{}, errors.New("multi-wallet deposit: concurrency must be greater than zero")
	}
	if cfg.Timeout <= 0 {
		return MultiWalletDepositResult{}, errors.New("multi-wallet deposit: timeout must be greater than zero")
	}
	if err := ctx.Err(); err != nil {
		return MultiWalletDepositResult{}, err
	}

	concurrency := min(cfg.Concurrency, len(cfg.Wallets))

	// 单次RPC请求结果
	type requestResult struct {
		operationID string        // 钱包操作唯一标识
		orderNo     string        // 存款订单号
		err         error         // 单次请求错误
		elapsed     time.Duration // 单次请求耗时
		started     bool          // 是否已发起请求
	}

	results := make([]requestResult, len(cfg.Wallets))
	jobs := make(chan int)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	startedAt := time.Now()

	// 启动固定数量的Worker处理不同钱包
	for range concurrency {
		go func() {
			defer wg.Done()

			for index := range jobs {
				w := cfg.Wallets[index]

				operationID := uuid.NewString()
				orderNo := "LOADTEST-" + uuid.NewString()

				results[index].operationID = operationID
				results[index].orderNo = orderNo
				results[index].started = true

				rpcCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
				requestStartedAt := time.Now()

				_, err := client.DepositCredit(rpcCtx, &depositpb.DepositCreditRequest{
					MemberGlobalId: w.MemberGlobalID, // 会员全局唯一标识
					OperationId:    operationID,      // 钱包操作唯一标识
					DepositOrderNo: orderNo,          // 存款订单号
					CurrencyCode:   w.CurrencyCode,   // 钱包币种
					Amount:         cfg.Amount,       // 单笔入账金额
				})

				elapsed := time.Since(requestStartedAt)
				cancel()

				results[index].err = err
				results[index].elapsed = elapsed
			}
		}()
	}

	// 分发钱包存款任务
	sending := true

	for i := range cfg.Wallets {
		if ctx.Err() != nil {
			break
		}

		select {
		case jobs <- i:
		case <-ctx.Done():
			sending = false
		}

		if !sending {
			break
		}
	}

	close(jobs)
	wg.Wait()

	elapsed := time.Since(startedAt)

	result := MultiWalletDepositResult{
		TotalRequests: len(cfg.Wallets),
		Elapsed:       elapsed,
	}

	// 汇总请求结果
	latencies := make([]time.Duration, 0, len(results))
	var totalLatency time.Duration

	for i, item := range results {
		if !item.started {
			result.SkippedRequests++
			continue
		}

		if item.err != nil {
			result.FailedRequests++
			result.Errors = append(result.Errors, fmt.Sprintf(
				"请求 %d: OperationID=%s, OrderNo=%s, 错误=%v",
				i+1,
				item.operationID,
				item.orderNo,
				item.err,
			))
			continue
		}

		result.SuccessRequests++
		latencies = append(latencies, item.elapsed)
		totalLatency += item.elapsed
	}

	// 统计成功请求的耗时分布
	if len(latencies) > 0 {
		slices.Sort(latencies)

		result.AvgLatency = totalLatency / time.Duration(len(latencies))
		result.P50Latency = latencyPercentile(latencies, 50)
		result.P95Latency = latencyPercentile(latencies, 95)
		result.P99Latency = latencyPercentile(latencies, 99)
		result.MaxLatency = latencies[len(latencies)-1]
	}

	// TPS按成功请求数计算
	if elapsed > 0 {
		result.TPS = float64(result.SuccessRequests) / elapsed.Seconds()
	}

	return result, nil
}

// latencyPercentile 获取已升序排列的请求耗时百分位值
func latencyPercentile(values []time.Duration, percent int) time.Duration {
	if len(values) == 0 {
		return 0
	}

	index := (len(values)*percent+99)/100 - 1

	return values[index]
}
