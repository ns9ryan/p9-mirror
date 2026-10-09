package operator

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-base/rpc/ent"
	operatorent "oa.98ent.com/p9/operator-base/rpc/ent/operator"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatoragentline"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatordomain"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorlanguage"
	"oa.98ent.com/p9/operator-base/rpc/ent/operatorregion"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// normalizeOperator 整理 operator 初始化基础数据
func normalizeOperator(req *InitializeRequest) {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.TimezoneCode = strings.TrimSpace(req.TimezoneCode)
	req.SettlementCurrencyCode = strings.TrimSpace(req.SettlementCurrencyCode)
}

// validateOperator 校验 operator 初始化基础数据
func validateOperator(req InitializeRequest) error {
	if req.Code == "" {
		return status.Error(codes.InvalidArgument, "operator code is required")
	}
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "operator name is required")
	}
	if req.TimezoneCode == "" {
		return status.Error(codes.InvalidArgument, "timezone code is required")
	}
	if req.SettlementCurrencyCode == "" {
		return status.Error(codes.InvalidArgument, "settlement currency code is required")
	}
	if req.Status < 1 || req.Status > 3 {
		return status.Error(codes.InvalidArgument, "operator status is invalid")
	}

	return nil
}

// getOperatorByCode 按业务编码查询 operator
func (s *Service) getOperatorByCode(ctx context.Context, code string) (*ent.Operator, bool, error) {
	data, err := s.db.Operator.
		Query().
		Where(operatorent.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("查询 operator 失败: %w", err)
	}

	return data, true, nil
}

// createOperator 创建 operator
func createOperator(ctx context.Context, tx *ent.Tx, req InitializeRequest) (*ent.Operator, error) {
	data, err := tx.Operator.
		Create().
		SetCode(req.Code).                                     // operator 全局唯一业务编码
		SetName(req.Name).                                     // operator 名称
		SetTimezoneCode(req.TimezoneCode).                     // IANA 时区编码
		SetSettlementCurrencyCode(req.SettlementCurrencyCode). // 结算货币编码
		SetStatus(req.Status).                                 // operator 状态: 1正常, 2暂停, 3关闭
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 operator 失败: %w", err)
	}

	return data, nil
}

// updateOperatorResources 更新已存在 operator 的资源配置（域名、语言、地区、代理线路）
func (s *Service) updateOperatorResources(ctx context.Context, operator *ent.Operator, req InitializeRequest) error {
	// 开启更新事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("开启更新事务失败: %w", err)
	}

	// 删除旧的域名
	if _, err = tx.OperatorDomain.Delete().Where(operatordomain.OperatorIDEQ(operator.ID)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除旧域名失败: %w", err)
	}

	// 删除旧的语言
	if _, err = tx.OperatorLanguage.Delete().Where(operatorlanguage.OperatorIDEQ(operator.ID)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除旧语言失败: %w", err)
	}

	// 删除旧的经营地区
	if _, err = tx.OperatorRegion.Delete().Where(operatorregion.OperatorIDEQ(operator.ID)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除旧经营地区失败: %w", err)
	}

	// 删除旧的代理子线路
	if _, err = tx.OperatorAgentLine.Delete().Where(operatoragentline.OperatorIDEQ(operator.ID)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除旧代理子线路失败: %w", err)
	}

	// 创建新的域名
	if err = createDomains(ctx, tx, operator.ID, req.Domains); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建新的语言
	if err = createLanguages(ctx, tx, operator.ID, req.LanguageCodes); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建新的经营地区
	if err = createRegions(ctx, tx, operator.ID, req.RegionCodes); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 创建新的代理子线路
	if err = createAgentLines(ctx, tx, operator.ID, req.AgentLineCodes); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 提交更新事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交更新事务失败: %w", err)
	}

	return nil
}
