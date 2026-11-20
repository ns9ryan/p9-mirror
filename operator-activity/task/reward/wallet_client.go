package reward

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/task/types"
)

// GrpcWalletClient gRPC钱包客户端实现
type GrpcWalletClient struct {
	logx.Logger
}

// 模拟钱包客户端实现
// NewGrpcWalletClient 创建gRPC钱包客户端
func NewGrpcWalletClient(ctx context.Context) types.WalletClient {
	return &GrpcWalletClient{
		Logger: logx.WithContext(ctx),
	}
}

// Payout 派发奖励
func (c *GrpcWalletClient) Payout(ctx context.Context, req *types.PayoutRequest) (*types.PayoutResponse, error) {
	c.Debugf("钱包派发奖励: wallet_order_no=%s, member_id=%d, amount_minor=%d",
		req.WalletOrderNo, req.MemberID, req.AmountMinor)
	// 转换响应
	payoutTime := time.Now()
	result := &types.PayoutResponse{
		Success:             true,
		Code:                "SUCCESS",
		Message:             "Mock payout success",
		WalletTransactionID: "mock_transaction_id",
		WalletOrderNo:       req.WalletOrderNo,
		PayoutAt:            payoutTime,
		ErrorCode:           "",
		ErrorMessage:        "",
		Retryable:           false,
	}

	c.Debugf("✓ 派发奖励成功: wallet_order_no=%s, transaction_id=%s",
		req.WalletOrderNo, "mock_transaction_id")

	return result, nil
}

// QueryPayoutStatus 查询派发状态
func (c *GrpcWalletClient) QueryPayoutStatus(ctx context.Context, walletOrderNo string) (*types.PayoutStatus, error) {
	c.Debugf("查询钱包派发状态: wallet_order_no=%s", walletOrderNo)

	// req := &memberwalletrpc.PayoutStatusRequest{
	// 	WalletOrderNo: walletOrderNo,
	// }

	// resp, err := c.client.QueryPayoutStatus(ctx, req)
	// if err != nil {
	// 	c.logger.Errorf("❌ 查询派发状态失败: wallet_order_no=%s, err=%v", walletOrderNo, err)
	// 	return nil, fmt.Errorf("query payout status failed: %w", err)
	// }

	payoutTime := time.Now()
	// if resp.PayoutAt != nil {
	// 	payoutTime = resp.PayoutAt.AsTime()
	// }

	updatedTime := time.Now()
	// if resp.UpdatedAt != nil {
	// 	updatedTime = resp.UpdatedAt.AsTime()
	// }

	return &types.PayoutStatus{
		Status:              "SUCCESS",
		AmountMinor:         0,
		Currency:            "USD",
		PayoutAt:            payoutTime,
		UpdatedAt:           updatedTime,
		WalletTransactionID: "mock_transaction_id_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		ErrorCode:           "",
	}, nil
}
