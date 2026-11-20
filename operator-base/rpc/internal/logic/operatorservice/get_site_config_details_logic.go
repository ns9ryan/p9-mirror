package operatorservicelogic

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorsitesetting"
	"oa.98ent.com/p9/operator-base/rpc/internal/config"
	operatorservice "oa.98ent.com/p9/operator-base/rpc/internal/operator"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/currencypb"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/regionpb"
	"oa.98ent.com/p9/platform-base/rpc/pb/platformbaserpc/timezonepb"
)

// GetSiteConfigDetailsLogic 查询站点配置详情及显示名称，不提供保存能力。
type GetSiteConfigDetailsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetSiteConfigDetailsLogic 创建配置详情查询逻辑。
func NewGetSiteConfigDetailsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSiteConfigDetailsLogic {
	return &GetSiteConfigDetailsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSiteConfigDetails 复用本地编码查询，再通过各自 RPC 补充名称及地区选项。
func (l *GetSiteConfigDetailsLogic) GetSiteConfigDetails(in *operatorpb.GetSiteConfigDetailsRequest) (
	*operatorpb.GetSiteConfigDetailsResponse, error,
) {
	operatorCode := ctxdata.OperatorCodeFromCtx(l.ctx)
	if operatorCode == "" {
		return nil, status.Error(codes.PermissionDenied, "login operator is required")
	}
	if l.svcCtx.CoreRpc == nil || l.svcCtx.RegionRpc == nil ||
		l.svcCtx.CurrencyRpc == nil || l.svcCtx.TimezoneRpc == nil {
		return nil, status.Error(codes.FailedPrecondition, "site config details RPC dependencies are not configured")
	}
	// 复用登录站点查询，读取本站已分配资源及勾选状态，不使用旧请求中的域名。
	config, owner, err := lookupSiteConfigWithOwner(l.ctx, l.svcCtx.DB, l.Logger)
	if err != nil {
		return nil, err
	}
	// 复用认证上下文查到的本站 ID，不接受客户端指定配置所属站点。
	settings, err := querySiteSettings(l.ctx, l.svcCtx.DB, owner, l.svcCtx.Config, l.Logger)
	if err != nil {
		return nil, err
	}
	result, err := l.loadDetails(l.ctx, config)
	if err != nil {
		l.Logger.Errorw("查询站点配置详情失败",
			logx.Field("operator_code", config.GetOperatorCode()), logx.Field("error", err.Error()))
		if l.ctx.Err() != nil {
			return nil, status.FromContextError(l.ctx.Err()).
				Err()
		}
		return nil, siteConfigDetailsError(err)
	}
	result.SiteSettings = settings
	return result, nil
}

// siteConfigDetailsError 保留下游状态及取消、超时语义，隐藏下游错误中的内部连接和实现细节。
func siteConfigDetailsError(err error) error {
	code := status.Code(err)
	if errors.Is(err, context.Canceled) {
		code = codes.Canceled
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = codes.DeadlineExceeded
	} else if code == codes.Unknown || code > codes.Unauthenticated {
		code = codes.Unavailable
	}
	return status.Error(code, "query site config details failed")
}

// loadDetails 返回已有配置的显示名称和经营地区选项，缺失资料保留编码，RPC 失败返回明确错误。
func (l *GetSiteConfigDetailsLogic) loadDetails(ctx context.Context, config *operatorpb.SiteConfigModel) (
	*operatorpb.GetSiteConfigDetailsResponse, error,
) {
	result := &operatorpb.GetSiteConfigDetailsResponse{SiteConfig: config}

	// 独立查询并行执行，使用同一个请求上下文，取消时停止全部下游请求。
	group, queryCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return l.loadLanguages(queryCtx, result)
	})
	group.Go(func() error {
		return l.loadRegions(queryCtx, result)
	})
	group.Go(func() error {
		return l.loadCurrency(queryCtx, result)
	})
	group.Go(func() error {
		return l.loadTimezone(queryCtx, result)
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// loadLanguages 查询本站 Core 语言资料，保留已停用和缺失的已分配语言。
func (l *GetSiteConfigDetailsLogic) loadLanguages(ctx context.Context, result *operatorpb.GetSiteConfigDetailsResponse) error {
	config := result.GetSiteConfig()
	languageRows, err := listSiteLanguages(ctx, l.svcCtx, config.GetOperatorCode())
	if err != nil {
		return err
	}
	// Core 只提供语言资料；最终展示范围以本站分配记录为准，避免返回未分配的语言。
	allocated := make([]string, 0, len(config.GetLanguageStatuses()))
	for code := range config.GetLanguageStatuses() {
		allocated = append(allocated, code)
	}
	if len(allocated) == 0 {
		allocated = append(allocated, config.GetLanguageCodes()...)
	}
	// 按编码稳定排序；本站勾选且 Core 未停用时才加入实际生效列表。
	sort.Strings(allocated)
	for _, code := range allocated {
		row := languageRows[code]
		if row != nil && config.GetLanguageStatuses()[code] == 1 && row.GetDisabled() == 0 {
			result.EffectiveLanguageCodes = append(result.EffectiveLanguageCodes, code)
		}
		// 将 Core RPC 返回值转换为本站定义的响应，资料缺失时仍保留已分配编码。
		result.Languages = append(result.Languages, &operatorpb.LanguageInfo{
			Id:        row.GetId(),
			Lang:      code,
			Disabled:  row.GetDisabled(),
			CreatedAt: row.GetCreatedAt(),
			UpdatedAt: row.GetUpdatedAt(),
			Name:      row.GetName(),
			SortNo:    row.GetSortNo(),
			I18NKey:   row.GetI18NKey(),
		})
	}
	return nil
}

// loadRegions 调用总网地区 RPC，组合本站选择与总网状态。
func (l *GetSiteConfigDetailsLogic) loadRegions(ctx context.Context, result *operatorpb.GetSiteConfigDetailsResponse) error {
	response, err := l.svcCtx.RegionRpc.ListAll(
		ctx,
		&regionpb.ListAllRegionsRequest{},
	)
	if err != nil {
		return fmt.Errorf("获取经营地区选项失败: %w", err)
	}
	if response == nil {
		return fmt.Errorf("获取经营地区选项返回空响应")
	}
	config := result.GetSiteConfig()
	regionStatuses := make(map[string]int64, len(config.GetRegionStatuses()))
	for code, regionStatus := range config.GetRegionStatuses() {
		regionStatuses[code] = regionStatus
	}
	// 总网资料只补充本站已分配地区的名称和状态；本站及总网都启用才实际生效。
	for _, row := range response.GetList() {
		if row == nil {
			continue
		}
		code := row.GetCode()
		regionStatus, exists := regionStatuses[code]
		if !exists {
			continue
		}
		effectiveStatus := int64(2)
		if regionStatus == 1 && row.GetStatus() == 1 {
			effectiveStatus = 1
			result.EffectiveRegionCodes = append(result.EffectiveRegionCodes, code)
		}
		// 将总网 RPC 资料逐字段转换为本站响应，不在本站 Proto 引用总网定义。
		result.RegionOptions = append(result.RegionOptions, &operatorpb.SiteRegionOption{
			Region: &operatorpb.RegionInfo{
				Id:          row.GetId(),
				Code:        row.GetCode(),
				CallingCode: row.GetCallingCode(),
				NameKey:     row.GetNameKey(),
				Status:      row.GetStatus(),
				SortNo:      row.GetSortNo(),
			},
			Status:          regionStatus,    // 启停状态
			EffectiveStatus: effectiveStatus, // 实际生效状态
		})
		delete(regionStatuses, code)
	}
	// 总网已缺失的资料仍保留本站分配编码及勾选状态，但实际生效状态固定为停用。
	missing := make([]string, 0, len(regionStatuses))
	for code := range regionStatuses {
		missing = append(missing, code)
	}
	sort.Strings(missing)
	for _, code := range missing {
		result.RegionOptions = append(result.RegionOptions, &operatorpb.SiteRegionOption{
			Region: &operatorpb.RegionInfo{
				Code: code, // 资料缺失时保留本站已分配编码
			},
			Status:          regionStatuses[code], // 启停状态
			EffectiveStatus: 2,                    // 实际生效状态
		})
	}
	return nil
}

// loadCurrency 调用总网货币 RPC，资料缺失时保留结算货币编码。
func (l *GetSiteConfigDetailsLogic) loadCurrency(ctx context.Context, result *operatorpb.GetSiteConfigDetailsResponse) error {
	code := result.GetSiteConfig().
		GetSettlementCurrencyCode()
	result.Currency = &operatorpb.CurrencyInfo{
		Code: code,
	}
	if code == "" {
		return nil
	}
	response, err := l.svcCtx.CurrencyRpc.GetByCode(
		ctx,
		&currencypb.GetCurrencyByCodeRequest{Code: code},
	)
	if status.Code(err) == codes.NotFound {
		// 未找到货币资料时仍返回先前设置的编码，不修改本站结算配置。
		return nil
	}
	if err != nil {
		return fmt.Errorf("获取货币失败: %w", err)
	}
	if response == nil {
		return fmt.Errorf("获取货币返回空响应")
	}
	if row := response.GetCurrency(); row != nil {
		// 总网 RPC 仍直接调用，仅将返回资料转换为本站响应结构。
		result.Currency = &operatorpb.CurrencyInfo{
			Id:           row.GetId(),
			Code:         row.GetCode(),
			NameKey:      row.GetNameKey(),
			CurrencyType: row.GetCurrencyType(),
			Symbol:       row.GetSymbol(),
			AmountFactor: row.GetAmountFactor(),
			Status:       row.GetStatus(),
			SortNo:       row.GetSortNo(),
		}
	}
	return nil
}

// loadTimezone 调用总网时区 RPC，资料缺失时保留时区编码。
func (l *GetSiteConfigDetailsLogic) loadTimezone(ctx context.Context, result *operatorpb.GetSiteConfigDetailsResponse) error {
	code := result.GetSiteConfig().
		GetTimezoneCode()
	result.Timezone = &operatorpb.TimezoneInfo{
		Code: code,
	}
	if code == "" {
		return nil
	}
	response, err := l.svcCtx.TimezoneRpc.GetByCode(
		ctx,
		&timezonepb.GetTimezoneByCodeRequest{Code: code},
	)
	if status.Code(err) == codes.NotFound {
		// 未找到时区资料时仍返回先前设置的编码，不修改本站时区配置。
		return nil
	}
	if err != nil {
		return fmt.Errorf("获取时区失败: %w", err)
	}
	if response == nil {
		return fmt.Errorf("获取时区返回空响应")
	}
	if row := response.GetTimezone(); row != nil {
		// 保持原时区字段和缺失资料回退规则，避免前端调用方式变化。
		result.Timezone = &operatorpb.TimezoneInfo{
			Id:      row.GetId(),
			Code:    row.GetCode(),
			NameKey: row.GetNameKey(),
			Status:  row.GetStatus(),
			SortNo:  row.GetSortNo(),
		}
	}
	return nil
}

// querySiteSettings 供完整资料查询按已校验的本站 ID 读取，未保存时只返回默认值，不插入记录。
func querySiteSettings(ctx context.Context, db *ent.Client, owner *ent.Operator, c config.Config, logger logx.Logger) (
	*operatorpb.SiteSettingsModel, error,
) {
	// 只读取已经核对归属的站点设置，不接受调用方另外指定站点主键。
	row, err := db.OperatorSiteSetting.Query().
		Where(operatorsitesetting.OperatorIDEQ(owner.ID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		// 老站尚未保存时仅在响应补标题、默认 Logo；其余字段保持零值，Configured 为 false。
		_, logoURL, err := operatorservice.LogoValue(c.DefaultSiteLogoURL, owner.Code, c)
		if err != nil {
			return nil, err
		}
		return &operatorpb.SiteSettingsModel{
			OperatorCode: owner.Code, // 分站编码
			SiteTitle:    owner.Name, // 网站标题
			SiteLogoUrl:  logoURL,    // 对外仍返回可直接使用的地址，不返回内部 Key
		}, nil
	}
	if err != nil {
		return nil, siteSettingsError(ctx, logger, err)
	}
	return toSiteSettingsModel(owner, row, c)
}
