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
	mode := flag.String("mode", "ping", "执行模式: ping, prepare, deposit")
	currencyCode := flag.String("currency-code", "USDT", "钱包结算币种")
	walletCount := flag.Int("wallets", 128, "测试钱包数量")
	memberIDStart := flag.Int64("member-id-start", 9000000001, "测试会员起始ID")
	partitionCount := flag.Int("partitions", 32, "Kafka Topic 分区数量")

	// 单笔资金操作测试
	memberGlobalID := flag.String("member-global-id", "", "测试会员全局唯一标识")
	amount := flag.Int64("amount", 0, "存款金额, 币种最小单位")

	flag.Parse()

	// 参数校验
	if *addr == "" {
		return fmt.Errorf("missing required argument: -addr")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	if *mode != "ping" && *mode != "prepare" && *mode != "deposit" {
		return fmt.Errorf("unsupported mode: %s", *mode)
	}
	if *partitionCount <= 0 {
		return fmt.Errorf("partition count must be greater than zero")
	}

	if *mode == "deposit" {
		if *memberGlobalID == "" {
			return fmt.Errorf("missing required argument: -member-global-id")
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
		fmt.Printf("OperationID:    %s\n", result.OperationID)
		fmt.Printf("OrderNo:        %s\n", result.OrderNo)
		fmt.Printf("BeforeBalance:  %d\n", result.BeforeBalance)
		fmt.Printf("AfterBalance:   %d\n", result.AfterBalance)
		fmt.Printf("BeforeVersion:  %d\n", result.BeforeVersion)
		fmt.Printf("AfterVersion:   %d\n", result.AfterVersion)
		fmt.Printf("Elapsed:        %s\n", result.Elapsed)

		return nil
	}

	// 创建测试钱包并预热 Actor
	startedAt := time.Now()

	wallets, err := setup.PrepareWallets(
		ctx,
		client.MemberWallet(),
		setup.WalletConfig{
			OperatorCode:  *operatorCode,  // 测试厅编码
			CurrencyCode:  *currencyCode,  // 钱包结算币种
			MemberIDStart: *memberIDStart, // 测试会员起始ID
			Count:         *walletCount,   // 测试钱包数量
			Timeout:       *timeout,       // 单次 RPC 请求超时时间
		},
	)
	if err != nil {
		return err
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
	fmt.Printf("准备耗时: %s\n", time.Since(startedAt))

	// 检查是否成功准备测试钱包
	if len(wallets) == 0 {
		return fmt.Errorf("no wallets prepared")
	}

	// 输出首个钱包信息, 方便手动核对
	first := wallets[0]

	fmt.Printf("\n首个测试钱包:\n")
	fmt.Printf("MemberID:       %d\n", first.MemberID)
	fmt.Printf("MemberGlobalID: %s\n", first.MemberGlobalID)
	fmt.Printf("WalletGlobalID: %s\n", first.WalletGlobalID)
	fmt.Printf("Balance:        %d\n", first.Balance)
	fmt.Printf("Version:        %d\n", first.Version)

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
