package operator

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-base/rpc/ent"
)

// InitializeRequest operator 初始化请求
type InitializeRequest struct {
	Code                   string             // operator 全局唯一业务编码
	Name                   string             // operator 名称
	TimezoneCode           string             // IANA 时区编码
	SettlementCurrencyCode string             // 结算货币编码
	Status                 int64              // operator 状态: 1正常, 2暂停, 3关闭
	Domains                []InitializeDomain // 当前有效域名
	LanguageCodes          []string           // 当前有效语言编码
	RegionCodes            []string           // 当前有效经营地区编码
	AgentLineCodes         []string           // 当前有效代理子线路编码
}

// Initialize 初始化 operator
func (s *Service) Initialize(ctx context.Context, req InitializeRequest) error {
	// 整理初始化数据
	normalizeInitializeRequest(&req)

	// 校验初始化数据
	if err := validateInitializeRequest(req); err != nil {
		return err
	}

	// 检查 operator 是否已经初始化
	existing, exists, err := s.getOperatorByCode(ctx, req.Code)
	if err != nil {
		return err
	}

	// 如果 operator 已存在, 则更新其基础资源配置
	if exists {
		return s.updateOperatorResources(ctx, existing, req)
	}

	// 先转换默认 Logo，再开启初始化事务；配置无效时不创建站点或写入默认设置。
	tx, logo, err := s.beginSiteInitialization(ctx, req.Code)
	if err != nil {
		return err
	}

	// 创建 operator 基础数据
	operatorData, err := createOperator(ctx, tx, req)
	if err != nil {
		_ = tx.Rollback()

		// 并发初始化相同 operator 时, 由 code 唯一约束完成最终幂等保护
		if ent.IsConstraintError(err) {
			existing, exists, queryErr := s.getOperatorByCode(ctx, req.Code)
			if queryErr != nil {
				return queryErr
			}
			if exists {
				return s.updateOperatorResources(ctx, existing, req)
			}
		}

		return err
	}

	// 仅新建站点时写入默认设置，与站点初始化一起提交；已有站点走上方资源更新分支。
	if err = createSiteSettings(ctx, tx, operatorData, logo); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建当前有效域名
	if err = createDomains(ctx, tx, operatorData.ID, req.Domains); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建当前有效语言
	if err = createLanguages(ctx, tx, operatorData.ID, req.LanguageCodes); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建当前有效经营地区
	if err = createRegions(ctx, tx, operatorData.ID, req.RegionCodes); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建当前有效代理子线路
	if err = createAgentLines(ctx, tx, operatorData.ID, req.AgentLineCodes); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 提交初始化事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交初始化事务失败: %w", err)
	}

	return nil
}

// beginSiteInitialization 校验默认 Logo 并开启建站事务，写入与提交仍由 Initialize 完成。
func (s *Service) beginSiteInitialization(ctx context.Context, operatorCode string) (*ent.Tx, string, error) {
	// 复用网站设置的 Logo 转换规则，已配置公开前缀时默认图片也保存 Key。
	logo, _, err := LogoValue(s.config.DefaultSiteLogoURL, operatorCode, s.config)
	if err != nil {
		return nil, "", fmt.Errorf("默认 Logo 配置无效: %w", err)
	}
	// 保持先校验配置再开启事务的顺序，不在这里执行写入、提交或回滚。
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("开启初始化事务失败: %w", err)
	}
	return tx, logo, nil
}

// normalizeInitializeRequest 整理 operator 初始化数据
func normalizeInitializeRequest(req *InitializeRequest) {
	normalizeOperator(req)
	normalizeDomains(req.Domains)
	normalizeLanguageCodes(req.LanguageCodes)
	normalizeRegionCodes(req.RegionCodes)
	normalizeAgentLineCodes(req.AgentLineCodes)
}

// validateInitializeRequest 校验 operator 初始化数据
func validateInitializeRequest(req InitializeRequest) error {
	// 校验 operator 基础数据
	if err := validateOperator(req); err != nil {
		return err
	}

	// 校验域名
	if err := validateDomains(req.Domains); err != nil {
		return err
	}

	// 校验语言
	if err := validateLanguageCodes(req.LanguageCodes); err != nil {
		return err
	}

	// 校验经营地区
	if err := validateRegionCodes(req.RegionCodes); err != nil {
		return err
	}

	// 校验代理子线路
	if err := validateAgentLineCodes(req.AgentLineCodes); err != nil {
		return err
	}

	return nil
}

// createSiteSettings 在初始化事务内写入默认设置，Logo 使用已转换的 Key 或兼容地址。
func createSiteSettings(ctx context.Context, tx *ent.Tx, owner *ent.Operator, logo string) error {
	// 标题取站点名称，Logo 取已转换的默认值；次数、验证码模式及修改人沿用表字段默认 0。
	if err := tx.OperatorSiteSetting.Create().
		SetOperatorID(owner.ID).
		SetSiteTitle(owner.Name).
		SetSiteLogoURL(logo).
		Exec(ctx); err != nil {
		return fmt.Errorf("初始化站点设置失败: %w", err)
	}
	return nil
}
