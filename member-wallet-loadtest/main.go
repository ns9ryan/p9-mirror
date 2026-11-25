package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"oa.98ent.com/p9/member-wallet-loadtest/internal/wallet"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", "", "钱包 RPC 服务地址")
	operatorCode := flag.String("operator-code", "OP_617E72", "测试厅编码")
	timeout := flag.Duration("timeout", 5*time.Second, "RPC 请求超时时间")

	flag.Parse()

	if *addr == "" {
		return fmt.Errorf("missing required argument: -addr")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	// 创建会员钱包 RPC 客户端
	client, err := wallet.NewClient(wallet.ClientConfig{
		Address:      *addr,         // RPC 服务地址
		OperatorCode: *operatorCode, // 测试厅编码
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

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return err
	}

	fmt.Println("member wallet RPC ping succeeded")
	return nil
}
