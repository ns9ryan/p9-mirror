package memberwalletlogic

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// withdrawalParams 提现资金操作参数
type withdrawalParams struct {
	memberGlobalID uuid.UUID // 会员全局唯一标识
	operationID    uuid.UUID // 钱包操作唯一标识
	withdrawalNo   string    // 提现订单号
	currencyCode   string    // 币种编码
	amount         int64     // 提现金额
}

// parseWithdrawalParams 校验并解析提现请求
func parseWithdrawalParams(
	memberGlobalIDText string,
	operationIDText string,
	withdrawalNo string,
	currencyCode string,
	amount int64,
) (withdrawalParams, error) {
	// 校验会员全局唯一标识
	memberGlobalID, err := uuid.Parse(memberGlobalIDText)
	if err != nil || memberGlobalID == uuid.Nil {
		return withdrawalParams{}, status.Error(codes.InvalidArgument, "会员全局标识无效")
	}

	// 校验钱包操作唯一标识
	operationID, err := uuid.Parse(operationIDText)
	if err != nil || operationID == uuid.Nil {
		return withdrawalParams{}, status.Error(codes.InvalidArgument, "钱包操作标识无效")
	}

	// 校验提现订单号
	if withdrawalNo == "" || strings.TrimSpace(withdrawalNo) != withdrawalNo || utf8.RuneCountInString(withdrawalNo) > 128 {
		return withdrawalParams{}, status.Error(codes.InvalidArgument, "提现订单号无效")
	}

	// 校验币种编码
	if currencyCode == "" || strings.TrimSpace(currencyCode) != currencyCode || utf8.RuneCountInString(currencyCode) > 16 {
		return withdrawalParams{}, status.Error(codes.InvalidArgument, "币种编码无效")
	}

	// 校验提现金额
	if amount <= 0 {
		return withdrawalParams{}, status.Error(codes.InvalidArgument, "提现金额必须大于零")
	}

	return withdrawalParams{
		memberGlobalID: memberGlobalID, // 会员全局唯一标识
		operationID:    operationID,    // 钱包操作唯一标识
		withdrawalNo:   withdrawalNo,   // 提现订单号
		currencyCode:   currencyCode,   // 币种编码
		amount:         amount,         // 提现金额
	}, nil
}
