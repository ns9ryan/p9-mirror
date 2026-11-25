package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

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
	mode := flag.String("mode", "ping", "执行模式: ping, prepare")
	currencyCode := flag.String("currency-code", "USDT", "钱包结算币种")
	walletCount := flag.Int("wallets", 128, "测试钱包数量")
	memberIDStart := flag.Int64("member-id-start", 9000000001, "测试会员起始ID")
	partitionCount := flag.Int("partitions", 32, "Kafka Topic 分区数量")

	flag.Parse()

	// 参数校验
	if *addr == "" {
		return fmt.Errorf("missing required argument: -addr")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	if *mode != "ping" && *mode != "prepare" {
		return fmt.Errorf("unsupported mode: %s", *mode)
	}
	if *partitionCount <= 0 {
		return fmt.Errorf("partition count must be greater than zero")
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
			Timeout:       *timeout,       // 单次RPC请求超时时间
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
	_, _ = fmt.Printf("\n测试钱包准备完成\n")
	_, _ = fmt.Printf("所属厅:   %s\n", *operatorCode)
	_, _ = fmt.Printf("钱包币种: %s\n", *currencyCode)
	_, _ = fmt.Printf("钱包数量: %d\n", len(wallets))
	_, _ = fmt.Printf("准备耗时: %s\n", time.Since(startedAt))

	// 输出首个钱包信息, 方便手动核对
	first := wallets[0]

	_, _ = fmt.Printf("\n首个测试钱包:\n")
	_, _ = fmt.Printf("MemberID:       %d\n", first.MemberID)
	_, _ = fmt.Printf("MemberGlobalID: %s\n", first.MemberGlobalID)
	_, _ = fmt.Printf("WalletGlobalID: %s\n", first.WalletGlobalID)
	_, _ = fmt.Printf("Balance:        %d\n", first.Balance)
	_, _ = fmt.Printf("Version:        %d\n", first.Version)

	// 输出 Partition 分布
	covered := 0
	missing := make([]int, 0)

	_, _ = fmt.Printf("\n目标 Partition 分布:\n")

	for index, count := range distribution {
		_, _ = fmt.Printf("Partition %02d: %d 个钱包\n", index, count)

		if count > 0 {
			covered++
		} else {
			missing = append(missing, index)
		}
	}

	_, _ = fmt.Printf("\n分区覆盖: %d/%d\n", covered, *partitionCount)

	if len(missing) > 0 {
		_, _ = fmt.Printf("未覆盖分区: %v\n", missing)
	}

	return nil
}
