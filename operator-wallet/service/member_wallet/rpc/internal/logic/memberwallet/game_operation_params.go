package memberwalletlogic

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gameOperationParams 游戏资金操作公共参数
type gameOperationParams struct {
	// 基础信息
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识

	// 业务信息
	businessNo   string // 游戏业务单号
	currencyCode string // 币种编码
}

// parseGameOperationParams 校验并解析游戏资金操作公共参数
func parseGameOperationParams(
	memberGlobalIDText string,
	operationIDText string,
	businessNo string,
	currencyCode string,
) (gameOperationParams, error) {
	// 校验会员全局唯一标识
	memberGlobalID, err := uuid.Parse(memberGlobalIDText)
	if err != nil || memberGlobalID == uuid.Nil {
		return gameOperationParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包操作唯一标识
	operationID, err := uuid.Parse(operationIDText)
	if err != nil || operationID == uuid.Nil {
		return gameOperationParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验游戏业务单号
	if businessNo == "" || strings.TrimSpace(businessNo) != businessNo || utf8.RuneCountInString(businessNo) > 128 {
		return gameOperationParams{}, status.Error(codes.InvalidArgument, "游戏业务单号无效")
	}

	// 校验币种编码
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return gameOperationParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	return gameOperationParams{
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识
		businessNo:     businessNo,     // 游戏业务单号
		currencyCode:   currencyCode,   // 币种编码
	}, nil
}

// validateGameAmount 校验游戏资金操作金额
func validateGameAmount(amount int64, allowZero bool) error {
	if amount < 0 || (!allowZero && amount == 0) {
		return status.Error(codes.InvalidArgument, "资金操作金额无效")
	}

	return nil
}
