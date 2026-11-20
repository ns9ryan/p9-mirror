package operator

import (
	"context"
	"net/http"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-base/api/internal/svc"
	"oa.98ent.com/p9/operator-base/api/internal/types"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

type GetSiteConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSiteConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSiteConfigLogic {
	return &GetSiteConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetSiteConfig 透传已校验的登录身份，查询当前登录站点详情。
func (l *GetSiteConfigLogic) GetSiteConfig() (*types.GetSiteConfigResponse, error) {
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" {
		return nil, xerr.Forbidden("login operator is required")
	}
	if l.svcCtx.OperatorRpc == nil {
		return nil, xerr.Err(http.StatusServiceUnavailable, "operator RPC is not configured")
	}
	// 调用本站 Operator RPC 的 GetSiteConfigDetails 方法。
	result, err := l.svcCtx.OperatorRpc.GetSiteConfigDetails(
		l.ctx,
		&operatorpb.GetSiteConfigDetailsRequest{},
	)
	if err != nil {
		return nil, err
	}
	// 返回资料再次核对登录站点，核对通过后才转换为前端响应。
	config := result.GetSiteConfig()
	if config == nil {
		return nil, xerr.InternalServerError("site config RPC returned no config")
	}
	if config.GetOperatorCode() != operatorCode {
		return nil, xerr.Forbidden("site config does not match login operator")
	}
	return toSiteConfigResponse(result), nil
}

// toSiteConfigResponse 转换 RPC 配置详情，复用设置转换并将空列表返回为 []。
func toSiteConfigResponse(result *operatorpb.GetSiteConfigDetailsResponse) *types.GetSiteConfigResponse {
	config := result.GetSiteConfig()
	currency := result.GetCurrency()
	timezone := result.GetTimezone()
	response := &types.GetSiteConfigResponse{
		SiteConfig: types.SiteConfig{
			OperatorCode:           config.GetOperatorCode(),                         // 分站编码
			DomainName:             config.GetDomainName(),                           // 后台域名
			LanguageCodes:          append([]string{}, config.GetLanguageCodes()...), // 本站勾选的语言编码
			RegionCodes:            append([]string{}, config.GetRegionCodes()...),   // 本站勾选的地区编码
			SettlementCurrencyCode: config.GetSettlementCurrencyCode(),               // 结算货币编码
			TimezoneCode:           config.GetTimezoneCode(),                         // IANA 时区编码
		},
		Languages:              make([]types.LanguageInfo, 0, len(result.GetLanguages())),         // 全部已分配语言
		RegionOptions:          make([]types.SiteRegionOption, 0, len(result.GetRegionOptions())), // 全部已分配地区
		EffectiveRegionCodes:   append([]string{}, result.GetEffectiveRegionCodes()...),           // 实际生效的地区编码
		EffectiveLanguageCodes: append([]string{}, result.GetEffectiveLanguageCodes()...),         // 实际生效的语言编码
		SiteSettings:           *toSiteSettingsInfo(result.GetSiteSettings()),                     // 本站设置及保存状态
		Currency: types.CurrencyInfo{
			Id:           currency.GetId(),           // 记录ID
			Code:         currency.GetCode(),         // 业务编码
			NameKey:      currency.GetNameKey(),      // 名称翻译 Key
			CurrencyType: currency.GetCurrencyType(), // 货币类型
			Symbol:       currency.GetSymbol(),       // 货币符号
			AmountFactor: currency.GetAmountFactor(), // 金额换算倍率
			Status:       currency.GetStatus(),       // 启停状态
			SortNo:       currency.GetSortNo(),       // 排序值
		},
		Timezone: types.TimezoneInfo{
			Id:      timezone.GetId(),      // 记录ID
			Code:    timezone.GetCode(),    // 业务编码
			NameKey: timezone.GetNameKey(), // 名称翻译 Key
			Status:  timezone.GetStatus(),  // 启停状态
			SortNo:  timezone.GetSortNo(),  // 排序值
		},
	}
	// 使用 RPC 返回的实际生效结果展示状态，保留本站原始勾选。
	effectiveLanguages := make(map[string]bool, len(result.GetEffectiveLanguageCodes()))
	for _, code := range result.GetEffectiveLanguageCodes() {
		effectiveLanguages[code] = true
	}
	for _, language := range result.GetLanguages() {
		languageStatus := config.GetLanguageStatuses()[language.GetLang()]
		effectiveStatus := int64(2)
		if effectiveLanguages[language.GetLang()] {
			effectiveStatus = 1
		}
		response.Languages = append(response.Languages, types.LanguageInfo{
			Status:          languageStatus,          // 启停状态
			EffectiveStatus: effectiveStatus,         // 实际生效状态
			Id:              language.GetId(),        // 记录ID
			Lang:            language.GetLang(),      // 语言编码
			Disabled:        language.GetDisabled(),  // Core 语言停用状态
			CreatedAt:       language.GetCreatedAt(), // 创建时间，Unix 秒
			UpdatedAt:       language.GetUpdatedAt(), // 更新时间，Unix 秒
			Name:            language.GetName(),      // 语言名称
			SortNo:          language.GetSortNo(),    // 排序值
			I18NKey:         language.GetI18NKey(),   // 语言名称翻译 Key
		})
	}
	for _, option := range result.GetRegionOptions() {
		region := option.GetRegion()
		response.RegionOptions = append(response.RegionOptions, types.SiteRegionOption{
			Region: types.RegionInfo{
				Id:          region.GetId(),          // 记录ID
				Code:        region.GetCode(),        // 业务编码
				CallingCode: region.GetCallingCode(), // 国际电话区号
				NameKey:     region.GetNameKey(),     // 名称翻译 Key
				Status:      region.GetStatus(),      // 启停状态
				SortNo:      region.GetSortNo(),      // 排序值
			},
			Status:          option.GetStatus(),          // 启停状态
			EffectiveStatus: option.GetEffectiveStatus(), // 实际生效状态
		})
	}
	return response
}
