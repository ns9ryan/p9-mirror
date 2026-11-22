package memberwalletlogic

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// rewardOperationParams 奖励资金操作公共参数
type rewardOperationParams struct {
	// 基础信息
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识

	// 业务信息
	businessNo   string // 奖励业务订单号
	currencyCode string // 币种编码
	amount       int64  // 奖励金额
}

// parseRewardOperationParams 校验并解析奖励资金操作参数
func parseRewardOperationParams(
	memberGlobalIDText string,
	operationIDText string,
	businessNo string,
	currencyCode string,
	amount int64,
) (rewardOperationParams, error) {
	// 校验会员全局唯一标识
	memberGlobalID, err := uuid.Parse(memberGlobalIDText)
	if err != nil || memberGlobalID == uuid.Nil {
		return rewardOperationParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包操作唯一标识
	operationID, err := uuid.Parse(operationIDText)
	if err != nil || operationID == uuid.Nil {
		return rewardOperationParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验奖励业务订单号
	if businessNo == "" || strings.TrimSpace(businessNo) != businessNo || utf8.RuneCountInString(businessNo) > 128 {
		return rewardOperationParams{}, status.Error(codes.InvalidArgument, "奖励业务订单号无效")
	}

	// 校验币种编码
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return rewardOperationParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验奖励金额
	if amount <= 0 {
		return rewardOperationParams{}, status.Error(codes.InvalidArgument, "奖励金额必须大于零")
	}

	return rewardOperationParams{
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识
		businessNo:     businessNo,     // 奖励业务订单号
		currencyCode:   currencyCode,   // 币种编码
		amount:         amount,         // 奖励金额
	}, nil
}
