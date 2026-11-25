package agent

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/currencypb"
)

// GetSiteCurrency 从分站配置读取结算货币编码及换算倍率，一次调用同时返回两者。
// 站点货币必须在初始化时配置；获取失败或货币编码为空直接返回错误。
func GetSiteCurrency(ctx context.Context, svcCtx *svc.ServiceContext) (currencyCode string, factor int64, err error) {
	opInfo, rpcErr := svcCtx.OperatorBaseRpc.GetOperatorInfo(ctx, nil)
	if rpcErr != nil {
		logx.WithContext(ctx).Errorf("[currency] GetOperatorInfo 失败: %v", rpcErr)
		return "", 0, xerr.Err(http.StatusServiceUnavailable, "site currency unavailable")
	}
	code := opInfo.GetSettlementCurrencyCode()
	if code == "" {
		return "", 0, xerr.Err(http.StatusInternalServerError, "site currency not configured")
	}
	f, fErr := GetCurrencyFactor(ctx, svcCtx, code)
	if fErr != nil {
		return "", 0, fErr
	}
	return code, f, nil
}

// GetSiteCurrencyFactor 仅需倍率时使用（已知货币编码，跳过 GetOperatorInfo）。
// currencyCode 必填；与分站结算货币不一致时直接返回错误。
func GetSiteCurrencyFactor(ctx context.Context, svcCtx *svc.ServiceContext, currencyCode string) (int64, error) {
	opInfo, rpcErr := svcCtx.OperatorBaseRpc.GetOperatorInfo(ctx, nil)
	if rpcErr != nil {
		logx.WithContext(ctx).Errorf("[currency] GetOperatorInfo 失败: %v", rpcErr)
		return 0, xerr.Err(http.StatusServiceUnavailable, "site currency unavailable")
	}
	siteCurrency := opInfo.GetSettlementCurrencyCode()
	if siteCurrency == "" {
		return 0, xerr.Err(http.StatusInternalServerError, "site currency not configured")
	}
	if currencyCode != siteCurrency {
		return 0, xerr.Err(http.StatusBadRequest, "currency mismatch: expected "+siteCurrency+", got "+currencyCode)
	}
	return GetCurrencyFactor(ctx, svcCtx, siteCurrency)
}

// GetCurrencyFactor 按货币编码查询换算倍率，查询失败或倍率无效返回错误。
func GetCurrencyFactor(ctx context.Context, svcCtx *svc.ServiceContext, currencyCode string) (int64, error) {
	if svcCtx.CurrencyRpc == nil {
		return 0, xerr.Err(http.StatusServiceUnavailable, "currency rpc not configured")
	}
	if currencyCode == "" {
		return 0, xerr.Err(http.StatusBadRequest, "currency code is empty")
	}
	resp, err := svcCtx.CurrencyRpc.GetByCode(ctx, &currencypb.GetCurrencyByCodeRequest{Code: currencyCode})
	if err != nil {
		logx.WithContext(ctx).Errorf("[currency] GetByCode(%s) 失败: %v", currencyCode, err)
		return 0, xerr.Err(http.StatusServiceUnavailable, "currency lookup failed")
	}
	if resp.GetCurrency() == nil || resp.GetCurrency().GetAmountFactor() <= 0 {
		return 0, xerr.Err(http.StatusInternalServerError, "invalid currency factor for "+currencyCode)
	}
	return resp.GetCurrency().GetAmountFactor(), nil
}

