package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/pb/memberwalletrpc/pingpb"
)

const operatorCodeHeader = "x-operator-code"

// ClientConfig 钱包 RPC 客户端配置
type ClientConfig struct {
	Address      string // RPC 服务地址
	OperatorCode string // 测试厅编码
}

// Client 会员钱包 RPC 客户端
type Client struct {
	conn         *grpc.ClientConn
	memberWallet memberwalletrpc.MemberWalletClient
	ping         memberwalletrpc.PingClient

	closeOnce sync.Once
	closeErr  error
}

// NewClient 创建会员钱包 RPC 客户端
func NewClient(cfg ClientConfig) (*Client, error) {
	address := strings.TrimSpace(cfg.Address)
	if address == "" {
		return nil, errors.New("wallet RPC address is required")
	}

	operatorCode := strings.TrimSpace(cfg.OperatorCode)
	if operatorCode == "" {
		return nil, errors.New("operator code is required")
	}

	// 当前内网测试环境使用明文 gRPC
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(operatorCodeInterceptor(operatorCode)),
	)
	if err != nil {
		return nil, fmt.Errorf("create wallet RPC client: %w", err)
	}

	return &Client{
		conn:         conn,
		memberWallet: memberwalletrpc.NewMemberWalletClient(conn),
		ping:         memberwalletrpc.NewPingClient(conn),
	}, nil
}

// Ping 检查钱包 RPC 服务是否可访问
func (c *Client) Ping(ctx context.Context) error {
	if ctx == nil {
		return errors.New("ping wallet RPC: nil context")
	}
	if c == nil || c.ping == nil {
		return errors.New("ping wallet RPC: client is not initialized")
	}

	_, err := c.ping.Ping(ctx, &pingpb.PingRequest{})
	if err != nil {
		return fmt.Errorf("ping wallet RPC: %w", err)
	}

	return nil
}

// MemberWallet 获取钱包资金服务客户端
func (c *Client) MemberWallet() memberwalletrpc.MemberWalletClient {
	if c == nil {
		return nil
	}

	return c.memberWallet
}

// Close 关闭 gRPC 连接
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}

	c.closeOnce.Do(func() {
		c.closeErr = c.conn.Close()
	})

	return c.closeErr
}

// operatorCodeInterceptor 为每个 RPC 请求附加厅编码
func operatorCodeInterceptor(operatorCode string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			operatorCodeHeader,
			operatorCode,
		)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
