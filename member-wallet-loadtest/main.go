package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"oa.98ent.com/p9/member-wallet-loadtest/internal/operation"
	"oa.98ent.com/p9/member-wallet-loadtest/internal/partition"
	"oa.98ent.com/p9/member-wallet-loadtest/internal/setup"
	"oa.98ent.com/p9/member-wallet-loadtest/internal/wallet"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// RPC 配置
	addr := flag.String("addr", "", "钱包 RPC 服务地址")
	operatorCode := flag.String("operator-code", "OP_034A80", "测试厅编码")
	timeout := flag.Duration("timeout", 5*time.Second, "单次 RPC 请求超时时间")

	// 测试配置
	mode := flag.String("mode", "ping", "执行模式: ping, prepare, deposit, concurrent, multi-wallet")
	currencyCode := flag.String("currency-code", "USDT", "钱包结算币种")
	walletCount := flag.Int("wallets", 128, "测试钱包数量")
	walletConcurrency := flag.Int("wallet-concurrency", 20, "钱包准备并发数")
	memberIDStart := flag.Int64("member-id-start", 9000000001, "测试会员起始ID")
	partitionCount := flag.Int("partitions", 32, "Kafka Topic 分区数量")

	// 资金操作测试
	memberGlobalID := flag.String("member-global-id", "", "测试会员全局唯一标识")
	amount := flag.Int64("amount", 0, "存款金额, 币种最小单位")
	concurrency := flag.Int("concurrency", 20, "并发请求数量")

	flag.Parse()

	// 参数校验
	if *addr == "" {
		return fmt.Errorf("missing required argument: -addr")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	if *mode != "ping" && *mode != "prepare" && *mode != "deposit" && *mode != "concurrent" && *mode != "multi-wallet" {
		return fmt.Errorf("unsupported mode: %s", *mode)
	}

	if *partitionCount <= 0 {
		return fmt.Errorf("partition count must be greater than zero")
	}

	// 单钱包资金操作参数校验
	if *mode == "deposit" || *mode == "concurrent" {
		if *memberGlobalID == "" {
			return fmt.Errorf("missing required argument: -member-global-id")
		}
		if *amount <= 0 {
			return fmt.Errorf("amount must be greater than zero")
		}
	}

	if *mode == "concurrent" && *concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than zero")
	}

	// 多钱包并发测试参数校验
	if *mode == "multi-wallet" {
		if *walletCount <= 0 {
			return fmt.Errorf("wallet count must be greater than zero")
		}
		if *walletConcurrency <= 0 {
			return fmt.Errorf("wallet concurrency must be greater than zero")
		}
		if *concurrency <= 0 {
			return fmt.Errorf("concurrency must be greater than zero")
		}
		if *amount <= 0 {
			return fmt.Errorf("amount must be greater than zero")
		}
	}

	// 创建会员钱包 RPC 客户端
	client, err := wallet.NewClient(wallet.ClientConfig{
		Address:      *addr,
		OperatorCode: *operatorCode,
	})
	if err != nil {
		return err
	}

	// 释放 gRPC 连接
	defer func() {
		if err := client.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "关闭钱包 RPC 连接失败: %v\n", err)
		}
	}()

	ctx := context.Background()

	// 执行测试前先检查 RPC 连通性
	pingCtx, cancel := context.WithTimeout(ctx, *timeout)
	err = client.Ping(pingCtx)
	cancel()

	if err != nil {
		return err
	}

	fmt.Println("会员钱包 RPC 连接测试成功")

	if *mode == "ping" {
		return nil
	}

	// 执行单笔存款测试
	if *mode == "deposit" {
		result, err := operation.Deposit(
			ctx,
			client.MemberWallet(),
			operation.DepositConfig{
				MemberGlobalID: *memberGlobalID, // 测试会员全局标识
				CurrencyCode:   *currencyCode,   // 钱包币种
				Amount:         *amount,         // 入账金额
				Timeout:        *timeout,        // 单次 RPC 超时时间
			},
		)
		if err != nil {
			return err
		}

		// 输出单笔存款测试结果
		fmt.Println("\n单笔存款测试成功")
		fmt.Printf("操作唯一标识:   %s\n", result.OperationID)
		fmt.Printf("存款订单号:     %s\n", result.OrderNo)
		fmt.Printf("操作前余额:     %d\n", result.BeforeBalance)
		fmt.Printf("操作后余额:     %d\n", result.AfterBalance)
		fmt.Printf("操作前版本:     %d\n", result.BeforeVersion)
		fmt.Printf("操作后版本:     %d\n", result.AfterVersion)
		fmt.Printf("本次测试耗时:   %s\n", result.Elapsed)

		return nil
	}

	// 执行同钱包并发存款测试
	if *mode == "concurrent" {
		result, err := operation.ConcurrentDeposit(
			ctx,
			client.MemberWallet(),
			operation.ConcurrentDepositConfig{
				MemberGlobalID: *memberGlobalID, // 测试会员全局标识
				CurrencyCode:   *currencyCode,   // 钱包币种
				Amount:         *amount,         // 单笔入账金额
				Concurrency:    *concurrency,    // 并发请求数量
				Timeout:        *timeout,        // 单次 RPC 超时时间
			},
		)
		if err != nil {
			return err
		}

		// 输出并发测试结果
		fmt.Println("\n同钱包并发存款测试结果")
		fmt.Printf("总请求数:       %d\n", result.TotalRequests)
		fmt.Printf("成功请求数:     %d\n", result.SuccessRequests)
		fmt.Printf("失败请求数:     %d\n", result.FailedRequests)

		fmt.Printf("操作前余额:     %d\n", result.BeforeBalance)
		fmt.Printf("操作后余额:     %d\n", result.AfterBalance)
		fmt.Printf("操作前版本:     %d\n", result.BeforeVersion)
		fmt.Printf("操作后版本:     %d\n", result.AfterVersion)

		fmt.Printf("余额校验通过:   %t\n", result.BalanceMatched)
		fmt.Printf("版本校验通过:   %t\n", result.VersionMatched)

		fmt.Printf("总耗时:         %s\n", result.Elapsed)
		fmt.Printf("平均请求耗时:   %s\n", result.AvgLatency)
		fmt.Printf("P95 请求耗时:   %s\n", result.P95Latency)
		fmt.Printf("最大请求耗时:   %s\n", result.MaxLatency)

		if len(result.Errors) > 0 {
			fmt.Println("\n失败请求:")
			for _, requestErr := range result.Errors {
				fmt.Println(requestErr)
			}
		}

		if result.FailedRequests > 0 {
			return fmt.Errorf(
				"%d concurrent requests failed; verify their OperationIDs and ledger records",
				result.FailedRequests,
			)
		}
		if !result.BalanceMatched || !result.VersionMatched {
			return fmt.Errorf("concurrent deposit result verification failed")
		}

		return nil
	}

	// 执行多钱包并发存款测试
	if *mode == "multi-wallet" {
		prepareStartedAt := time.Now()

		// 准备不同会员的钱包并预热 Actor
		wallets, err := setup.PrepareWallets(
			ctx,
			client.MemberWallet(),
			setup.WalletConfig{
				OperatorCode:  *operatorCode,      // 测试厅编码
				CurrencyCode:  *currencyCode,      // 钱包币种
				MemberIDStart: *memberIDStart,     // 测试会员起始ID
				Count:         *walletCount,       // 测试钱包数量
				Concurrency:   *walletConcurrency, // 钱包准备并发数
				Timeout:       *timeout,           // 单次 RPC 请求超时时间
			},
		)
		if err != nil {
			return err
		}

		prepareElapsed := time.Since(prepareStartedAt)

		if len(wallets) != *walletCount {
			return fmt.Errorf("prepared wallet count mismatch: expected=%d actual=%d", *walletCount, len(wallets))
		}

		// 提取钱包全局标识
		walletGlobalIDs := make([]string, len(wallets))
		for i, w := range wallets {
			walletGlobalIDs[i] = w.WalletGlobalID
		}

		// 统计钱包的 Partition 分布
		distribution, err := partition.Distribution(walletGlobalIDs, *partitionCount)
		if err != nil {
			return fmt.Errorf("统计钱包 Partition 分布失败: %w", err)
		}

		// 输出钱包准备结果
		fmt.Println("\n测试钱包准备完成")
		fmt.Printf("所属厅:         %s\n", *operatorCode)
		fmt.Printf("钱包币种:       %s\n", *currencyCode)
		fmt.Printf("钱包数量:       %d\n", len(wallets))
		fmt.Printf("准备并发数:     %d\n", *walletConcurrency)
		fmt.Printf("准备耗时:       %s\n", prepareElapsed)

		// 输出 Partition 分布
		covered := 0

		fmt.Println("\n钱包 Partition 分布:")

		for index, count := range distribution {
			fmt.Printf("Partition %02d: %d 个钱包\n", index, count)

			if count > 0 {
				covered++
			}
		}

		fmt.Printf("分区覆盖:       %d/%d\n", covered, *partitionCount)

		// 正式开始多钱包并发测试
		fmt.Println("\n开始多钱包并发存款测试")

		result, err := operation.MultiWalletDeposit(
			ctx,
			client.MemberWallet(),
			operation.MultiWalletDepositConfig{
				Wallets:     wallets,      // 测试钱包
				Amount:      *amount,      // 单笔存款金额
				Concurrency: *concurrency, // 最大并发请求数
				Timeout:     *timeout,     // 单次 RPC 超时时间
			},
		)
		if err != nil {
			return err
		}

		// 输出多钱包并发测试结果
		fmt.Println("\n多钱包并发存款测试结果")
		fmt.Printf("总钱包数:       %d\n", result.TotalRequests)
		fmt.Printf("成功请求数:     %d\n", result.SuccessRequests)
		fmt.Printf("失败请求数:     %d\n", result.FailedRequests)
		fmt.Printf("未执行请求数:   %d\n", result.SkippedRequests)

		fmt.Printf("最大并发数:     %d\n", min(*concurrency, len(wallets)))

		fmt.Printf("总耗时:         %s\n", result.Elapsed)
		fmt.Printf("实际 TPS:       %.2f\n", result.TPS)

		fmt.Printf("平均请求耗时:   %s\n", result.AvgLatency)
		fmt.Printf("P50 请求耗时:   %s\n", result.P50Latency)
		fmt.Printf("P95 请求耗时:   %s\n", result.P95Latency)
		fmt.Printf("P99 请求耗时:   %s\n", result.P99Latency)
		fmt.Printf("最大请求耗时:   %s\n", result.MaxLatency)

		if len(result.Errors) > 0 {
			fmt.Println("\n失败请求:")
			for _, requestErr := range result.Errors {
				fmt.Println(requestErr)
			}
		}

		if result.FailedRequests > 0 || result.SkippedRequests > 0 {
			return fmt.Errorf(
				"multi-wallet test incomplete: failed=%d skipped=%d; verify failed requests by OperationID and OrderNo",
				result.FailedRequests,
				result.SkippedRequests,
			)
		}

		return nil
	}

	// 创建测试钱包并预热 Actor
	startedAt := time.Now()

	wallets, err := setup.PrepareWallets(
		ctx,
		client.MemberWallet(),
		setup.WalletConfig{
			OperatorCode:  *operatorCode,      // 测试厅编码
			CurrencyCode:  *currencyCode,      // 钱包币种
			MemberIDStart: *memberIDStart,     // 测试会员起始ID
			Count:         *walletCount,       // 测试钱包数量
			Concurrency:   *walletConcurrency, // 钱包准备并发数
			Timeout:       *timeout,           // 单次 RPC 请求超时时间
		},
	)
	if err != nil {
		return err
	}

	// 检查是否成功准备测试钱包
	if len(wallets) == 0 {
		return fmt.Errorf("no wallets prepared")
	}

	// 提取钱包全局标识
	walletGlobalIDs := make([]string, len(wallets))
	for i, w := range wallets {
		walletGlobalIDs[i] = w.WalletGlobalID
	}

	// 计算各 Partition 的钱包分布
	distribution, err := partition.Distribution(walletGlobalIDs, *partitionCount)
	if err != nil {
		return fmt.Errorf("统计钱包 Partition 分布失败: %w", err)
	}

	// 输出测试数据准备结果
	fmt.Printf("\n测试钱包准备完成\n")
	fmt.Printf("所属厅:   %s\n", *operatorCode)
	fmt.Printf("钱包币种: %s\n", *currencyCode)
	fmt.Printf("钱包数量: %d\n", len(wallets))
	fmt.Printf("准备并发数: %d\n", *walletConcurrency)
	fmt.Printf("准备耗时: %s\n", time.Since(startedAt))

	// 输出首个钱包信息, 方便手动核对
	first := wallets[0]

	fmt.Printf("\n首个测试钱包:\n")
	fmt.Printf("会员ID:         %d\n", first.MemberID)
	fmt.Printf("会员全局标识:   %s\n", first.MemberGlobalID)
	fmt.Printf("钱包全局标识:   %s\n", first.WalletGlobalID)
	fmt.Printf("钱包余额:       %d\n", first.Balance)
	fmt.Printf("钱包版本:       %d\n", first.Version)

	// 输出 Partition 分布
	covered := 0
	missing := make([]int, 0)

	fmt.Printf("\n目标 Partition 分布:\n")

	for index, count := range distribution {
		fmt.Printf("Partition %02d: %d 个钱包\n", index, count)

		if count > 0 {
			covered++
		} else {
			missing = append(missing, index)
		}
	}

	fmt.Printf("\n分区覆盖: %d/%d\n", covered, *partitionCount)

	if len(missing) > 0 {
		fmt.Printf("未覆盖分区: %v\n", missing)
	}

	return nil
}
