package operatorservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/core/rpc/coreclient"
	"oa.98ent.com/p9/operator-base/rpc/ent"
	"oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatordomain"
	"oa.98ent.com/p9/operator-base/rpc/internal/config"
	operatorservice "oa.98ent.com/p9/operator-base/rpc/internal/operator"
	"oa.98ent.com/p9/operator-base/rpc/internal/svc"
	"oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// normalizeBackendDomain 统一后台域名的大小写并去除末尾的点。
func normalizeBackendDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
}

// lookupSiteOwner 供配置查询、保存和勾选操作共用，仅从认证上下文确定所属站点。
// db 可以是事务客户端，写操作在事务内复用时仍按同一登录站点核对归属。
func lookupSiteOwner(ctx context.Context, db *ent.Client, logger logx.Logger) (*ent.Operator, error) {
	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
	if operatorCode == "" {
		return nil, status.Error(codes.PermissionDenied, "login operator is required")
	}
	if db == nil {
		return nil, status.Error(codes.FailedPrecondition, "operator database is not configured")
	}
	// 只查询登录身份对应的站点，忽略 HTTP Host、请求参数及旧 RPC 的域名字段。
	owner, err := db.Operator.Query().
		Where(operator.CodeEQ(operatorCode)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "login operator not found")
	}
	if err != nil {
		return nil, siteSettingsError(ctx, logger, err)
	}
	if owner.Code != operatorCode {
		return nil, status.Error(codes.PermissionDenied, "site does not match login operator")
	}
	return owner, nil
}

// lookupSiteConfigWithOwner 读取登录站点的配置及分配记录，后台域名仅用于响应展示。
func lookupSiteConfigWithOwner(ctx context.Context, db *ent.Client, logger logx.Logger) (*operatorpb.SiteConfigModel, *ent.Operator, error) {
	owner, err := lookupSiteOwner(ctx, db, logger)
	if err != nil {
		return nil, nil, err
	}
	// 只读取本站已分配的语言和地区，保持原勾选与实际生效的业务含义。
	owner.Edges.Languages, err = owner.QueryLanguages().
		All(ctx)
	if err != nil {
		return nil, nil, siteSettingsError(ctx, logger, err)
	}
	owner.Edges.Regions, err = owner.QueryRegions().
		All(ctx)
	if err != nil {
		return nil, nil, siteSettingsError(ctx, logger, err)
	}
	// 后台域名从本站数据库读取；未配置域名时返回空字符串，不阻断登录后配置管理。
	domain, err := owner.QueryDomains().
		Where(operatordomain.DomainTypeEQ(domainTypeBackend)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, nil, siteSettingsError(ctx, logger, err)
	}
	domainName := ""
	if domain != nil {
		domainName = domain.DomainName
	}
	return toSiteConfigModel(domainName, owner), owner, nil
}

// toSiteConfigModel 将本地站点及分配记录转换为响应，勾选编码仅包含状态 1，状态表保留全部分配项。
func toSiteConfigModel(domainName string, operatorData *ent.Operator) *operatorpb.SiteConfigModel {
	// 状态表保留全部已分配语言；勾选列表只收集状态 1，实际生效状态另结合 Core 资料判断。
	languageCodes := make([]string, 0, len(operatorData.Edges.Languages))
	languageStatuses := make(map[string]int64, len(operatorData.Edges.Languages))
	for _, item := range operatorData.Edges.Languages {
		languageStatuses[item.LanguageCode] = item.Status
		if item.Status == 1 {
			languageCodes = append(languageCodes, item.LanguageCode)
		}
	}

	// 地区同样只从本站分配记录取选项，不把总网全部地区直接当作本站可选项。
	regionCodes := make([]string, 0, len(operatorData.Edges.Regions))
	regionStatuses := make(map[string]int64, len(operatorData.Edges.Regions))
	for _, item := range operatorData.Edges.Regions {
		regionStatuses[item.RegionCode] = item.Status
		if item.Status == 1 {
			regionCodes = append(regionCodes, item.RegionCode)
		}
	}

	return &operatorpb.SiteConfigModel{
		OperatorCode:           operatorData.Code,                   // 分站编码
		DomainName:             domainName,                          // 数据库中的后台域名，未配置时为空
		LanguageCodes:          languageCodes,                       // 本站勾选的语言编码
		LanguageStatuses:       languageStatuses,                    // 本站全部已分配语言及各自启停状态
		RegionCodes:            regionCodes,                         // 本站勾选的地区编码
		RegionStatuses:         regionStatuses,                      // 本站全部已分配地区及各自启停状态
		SettlementCurrencyCode: operatorData.SettlementCurrencyCode, // 结算货币编码
		TimezoneCode:           operatorData.TimezoneCode,           // IANA 时区编码
	}
}

// toSiteSettingsModel 转换本站设置，将内部 Logo Key 还原为公开地址，时间使用 Unix 秒。
func toSiteSettingsModel(owner *ent.Operator, row *ent.OperatorSiteSetting, c config.Config) (*operatorpb.SiteSettingsModel, error) {
	// 复用 Logo 转换规则：Key 按公开前缀拼接，兼容地址按原值返回；此处不回写数据库。
	_, logoURL, err := operatorservice.LogoValue(row.SiteLogoURL, owner.Code, c)
	if err != nil {
		return nil, err
	}
	// 原样保留已保存的标题和限制值，不因字段为空或为 0 再覆盖成初始化默认值。
	result := &operatorpb.SiteSettingsModel{
		OperatorCode: owner.Code, // 已核对归属的分站编码
	}
	result.SiteTitle = row.SiteTitle                               // 已保存的网站标题
	result.SiteLogoUrl = logoURL                                   // 转换后的 Logo 展示地址，允许为空
	result.EmailLoginLimit = row.EmailLoginLimit                   // 邮箱登录阈值，0 不限制
	result.SmsLoginLimit = row.SmsLoginLimit                       // 短信登录阈值，0 不限制
	result.MemberPasswordErrorLimit = row.MemberPasswordErrorLimit // 密码错误次数上限，0 不限制
	result.IpMemberRegisterLimit = row.IPMemberRegisterLimit       // 同 IP 注册会员数上限，0 不限制
	result.RegisterCaptchaMode = row.RegisterCaptchaMode           // 注册验证：0 不验证、1 拖拽、2 图形
	result.LoginCaptchaMode = row.LoginCaptchaMode                 // 登录验证：0 不验证、1 拖拽、2 图形
	// 审计信息来自已保存记录；Configured 为 true 表示确有记录，不由字段是否为 0 判断。
	result.UpdatedBy = row.UpdatedBy        // 已保存的修改人 ID，初始化时为 0
	result.CreatedAt = row.CreatedAt.Unix() // 创建时间，Unix 秒
	result.UpdatedAt = row.UpdatedAt.Unix() // 修改时间，Unix 秒
	result.Configured = true                // 已存在设置记录，与限制值是否为 0 无关
	return result, nil
}

// siteSettingsError 复用调用方的 Logger 记录数据库错误，保留并发冲突及内部错误转换。
func siteSettingsError(ctx context.Context, logger logx.Logger, err error) error {
	logger.Errorw("站点设置数据库操作失败", logx.Field("error", err.Error()))
	if ctx.Err() != nil {
		return status.FromContextError(ctx.Err()).
			Err()
	}
	var pgErr *pgconn.PgError
	// PostgreSQL 40001 为序列化冲突，40P01 为死锁；返回 Aborted，要求重新读取后重试。
	if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
		return status.Error(codes.Aborted, "site settings changed concurrently; reload and retry")
	}
	// 未知数据库错误只向调用方返回通用消息，不暴露 SQL 或内部连接信息。
	return status.Error(codes.Internal, "site settings operation failed")
}

const (
	// 单轮最多查询 100 页；分页快照变化时最多重新读取一轮，合计两轮。
	maxLanguagePages    int32 = 100
	maxLanguageAttempts       = 2
)

var errLanguagePaginationChanged = errors.New("本站语言分页数据发生变化，请重试")

// listSiteLanguages 按本站身份获取全部分页，不传 disabled，保留已停用语言的名称。
func listSiteLanguages(ctx context.Context, svcCtx *svc.ServiceContext, operatorCode string) (map[string]*coreclient.I18NLangInfo, error) {
	// Core 返回的语言资料用于补充名称和状态；详情查询仍以本地分配记录决定本站选项。
	// WithOperatorCode 会追加 metadata；先移除旧站点值，确保 Core 按本次登录站点过滤。
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		md = md.Copy()
		md.Delete("x-operator-code")
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	ctx = ctxdata.WithOperatorCode(ctx, operatorCode)
	// 仅分页快照不一致时重新读取，最多两轮；下游 RPC 错误直接返回，不自动重试。
	for attempt := 0; attempt < maxLanguageAttempts; attempt++ {
		rows, err := readLanguagePages(ctx, svcCtx)
		if !errors.Is(err, errLanguagePaginationChanged) {
			return rows, err
		}
	}
	// 两轮均不一致时返回错误，不把任一轮的部分结果当作完整语言列表。
	return nil, errLanguagePaginationChanged
}

// readLanguagePages 读取全部语言分页，发现重复、缺页或总数变化时交给上层重试。
func readLanguagePages(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]*coreclient.I18NLangInfo, error) {
	rows := make(map[string]*coreclient.I18NLangInfo)
	var total int64
	for page := int32(1); page <= maxLanguagePages; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// 调用本站 Core RPC，每页最多 1000 条、最多 100 页；不筛掉停用项，供查询展示和新启用校验。
		response, err := svcCtx.CoreRpc.GetI18NLangList(
			ctx,
			&coreclient.I18NLangListReq{
				Page:     page,
				PageSize: 1000,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("获取本站语言列表失败: %w", err)
		}
		if response == nil {
			return nil, fmt.Errorf("获取本站语言列表返回空响应")
		}
		// 本轮分页总数必须稳定，重复编码、提前空页或累计条数异常均视为快照变化。
		if page == 1 {
			total = response.GetTotal()
		}
		if total < 0 || response.GetTotal() != total {
			return nil, errLanguagePaginationChanged
		}
		previousCount := len(rows)
		for _, row := range response.GetList() {
			if row == nil || row.GetLang() == "" || rows[row.GetLang()] != nil {
				return nil, errLanguagePaginationChanged
			}
			rows[row.GetLang()] = row
		}
		if int64(len(rows)) > total {
			return nil, errLanguagePaginationChanged
		}
		if int64(len(rows)) == total {
			// 唯一编码数量达到总数才返回；总数为 0 时返回已初始化的空 map。
			return rows, nil
		}
		if len(rows) == previousCount {
			// 尚未达到总数却没有新增记录，说明出现提前空页，不能继续拼出不完整结果。
			return nil, errLanguagePaginationChanged
		}
	}
	// 超过单轮页数上限直接失败，避免无限分页或返回被截断的资料。
	return nil, fmt.Errorf("本站语言分页超出范围")
}
