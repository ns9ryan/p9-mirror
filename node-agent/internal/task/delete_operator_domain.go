package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	operatorpb "oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// deleteOperatorDomainParams 删除分站域名任务参数
type deleteOperatorDomainParams struct {
	OperatorCode string   `json:"operator_code"`
	DomainNames  []string `json:"domain_names"`
}

// executeDeleteOperatorDomain 执行删除分站域名任务
func (s *Service) executeDeleteOperatorDomain(ctx context.Context, rawParams json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)

	// ==================== 第一步：解析参数 ====================

	var params deleteOperatorDomainParams
	if err := json.Unmarshal(rawParams, &params); err != nil {
		logger.Errorf("解析删除域名参数失败: error=%v", err)
		return nil, fmt.Errorf("解析删除域名参数失败: %w", err)
	}

	operatorCode := params.OperatorCode
	if operatorCode == "" {
		logger.Errorf("分站编码为空")
		return nil, fmt.Errorf("分站编码为空")
	}

	logger.Infof("开始执行删除分站域名任务: operator_code=%s", operatorCode)

	// ==================== 第一步：获取分站所有域名列表 ====================
	// 获取域名列表
	var domainNames []string = params.DomainNames

	// ==================== 第二步：调用 operator-base 删除分站域名 ====================

	deleteResp, err := s.operatorBaseRpc.DeleteOperatorDomains(ctx, &operatorpb.DeleteOperatorDomainsRequest{
		OperatorCode: operatorCode,
		DomainNames:  domainNames,
	})
	if err != nil {
		logger.Errorf("调用operator-base.DeleteOperatorDomains失败: operator_code=%s, error=%v", operatorCode, err)
		return nil, fmt.Errorf("删除分站域名失败: %w", err)
	}

	if deleteResp == nil {
		logger.Errorf("operator-base返回删除域名响应为空: operator_code=%s", operatorCode)
		return nil, fmt.Errorf("operator-base返回删除域名响应为空")
	}

	if !deleteResp.Success {
		logger.Errorf("删除分站域名失败: operator_code=%s, message=%s", operatorCode, deleteResp.Message)
		return nil, fmt.Errorf("删除分站域名失败: %s", deleteResp.Message)
	}

	logger.Infof("删除分站域名成功: operator_code=%s, 删除域名数=%d",
		operatorCode,
		deleteResp.DomainDeleted,
	)

	// ==================== 第三步：删除 nginx 本地配置文件 ====================

	// 如果 NginxMgr 已配置，则删除本地配置文件
	if s.DomainManager != nil && len(domainNames) > 0 {
		logger.Infof("开始删除 nginx 本地配置: operator_code=%s, domain_count=%d", operatorCode, len(domainNames))

		// 逐个删除每个域名的配置文件
		for _, domain := range domainNames {
			if err := s.DomainManager.DeleteConfig(domain); err != nil {
				logger.Errorf("删除 nginx 配置失败（继续处理）: domain=%s, error=%v", domain, err)
			} else {
				logger.Infof("nginx 配置已删除: domain=%s", domain)
			}
		}

		// ==================== 第四步：重载 nginx ====================

		logger.Infof("调用 nginx reload 重新加载配置: operator_code=%s", operatorCode)
		reloadOutput, err := s.DomainManager.ReloadNginx()
		if err != nil {
			logger.Errorf("nginx reload 失败（继续处理）: operator_code=%s, error=%v, output=%s", operatorCode, err, reloadOutput)
			// 继续处理，不中断流程
		} else {
			logger.Infof("nginx 重载成功: operator_code=%s, output=%s", operatorCode, reloadOutput)
		}
	} else if len(domainNames) == 0 {
		logger.Infof("没有域名被删除，跳过 nginx 配置清理: operator_code=%s", operatorCode)
	}

	return nil, nil
}
