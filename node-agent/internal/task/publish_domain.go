package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	operatorbasepb "oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
	platformoperatorpb "oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"
)

// publishDomainParams 发布域名任务参数
type publishDomainParams struct {
	OperatorCode string `json:"operator_code"` // 分站全局唯一业务编码
}

// executePublishDomain 执行发布域名任务
func (s *Service) executePublishDomain(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	logx.WithContext(ctx).Infof("开始执行发布域名任务: params=%s", string(params))

	// 解析任务参数
	var publishDomainParam publishDomainParams
	if err := json.Unmarshal(params, &publishDomainParam); err != nil {
		errMsg := fmt.Sprintf("解析发布域名任务参数失败: %v", err)
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	operatorCode := strings.TrimSpace(publishDomainParam.OperatorCode)
	if operatorCode == "" {
		errMsg := "分站编码不能为空"
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	logx.WithContext(ctx).Infof("开始发布分站域名, operator_code=%s", operatorCode)

	// ==================== 第一步：从 platform-operator 获取分站域名 ====================

	logx.WithContext(ctx).Debugf("准备从platform-operator获取分站域名, operator_code=%s", operatorCode)

	initDataResp, err := s.platformOperatorRpc.GetInitializationData(ctx, &platformoperatorpb.GetInitializationDataRequest{
		OperatorCode: operatorCode,
	})
	if err != nil {
		errMsg := fmt.Sprintf("调用platform-operator.GetInitializationData失败: operator_code=%s, error=%v", operatorCode, err)
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if initDataResp == nil {
		errMsg := fmt.Sprintf("platform-operator返回初始化数据为空: operator_code=%s", operatorCode)
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	logx.WithContext(ctx).Infof("成功获取分站初始化数据, operator_code=%s, 域名数量=%d", operatorCode, len(initDataResp.Domains))

	// ==================== 第二步：调用 operator-base 初始化域名 ====================

	logx.WithContext(ctx).Infof("准备调用operator-base.InitializeDomains, operator_code=%s, 域名数量=%d", operatorCode, len(initDataResp.Domains))

	initDomainsResp, err := s.operatorBaseRpc.InitializeDomains(ctx, &operatorbasepb.InitializeDomainsRequest{
		OperatorCode: operatorCode,
		DomainList:   convertToDomainModels(initDataResp.Domains),
	})
	if err != nil {
		errMsg := fmt.Sprintf("调用operator-base.InitializeDomains失败: operator_code=%s, error=%v", operatorCode, err)
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if initDomainsResp == nil {
		errMsg := fmt.Sprintf("operator-base返回初始化域名响应为空: operator_code=%s", operatorCode)
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if !initDomainsResp.Success {
		errMsg := fmt.Sprintf("初始化分站域名失败: operator_code=%s, message=%s", operatorCode, initDomainsResp.Message)
		logx.WithContext(ctx).Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	logx.WithContext(ctx).Infof("分站域名发布成功, operator_code=%s, 初始化统计: 成功=%d", operatorCode, initDomainsResp.DomainStat.Success)

	// ==================== 新增：生成 Nginx 配置 ====================

	// 为每个域名生成 nginx 配置
	for _, domain := range initDataResp.Domains {
		// 只处理分站后台域名（DomainType=1）
		if domain == nil || domain.DomainName == "" || domain.DomainType != 1 {
			continue
		}

		if err := s.DomainManager.GenerateConfig(domain.DomainName); err != nil {
			logx.WithContext(ctx).Errorf("生成 nginx 配置失败: domain=%s, error=%v",
				domain.DomainName, err)
			// 不中断流程，继续处理其他域名
		}
	}

	// 验证并重载 nginx（所有配置生成后统一执行）
	if s.DomainManager != nil {
		if reloadOutput, err := s.DomainManager.ReloadNginx(); err != nil {
			logx.WithContext(ctx).Errorf("Nginx 重载失败: error=%v, output=%s", err, reloadOutput)
			// 不中断流程，记录错误即可
		}
	}

	logx.WithContext(ctx).Infof("分站域名发布和 nginx 配置完成: operator_code=%s", operatorCode)
	return json.RawMessage(`{"success":true}`), nil
}

// convertToDomainModels 将platform-operator的域名转换为operator-base的域名模型
func convertToDomainModels(sourceList []*platformoperatorpb.DomainInitializationInfo) []*operatorbasepb.DomainModel {
	if len(sourceList) == 0 {
		return nil
	}

	var result []*operatorbasepb.DomainModel
	for _, src := range sourceList {
		if src == nil {
			continue
		}
		result = append(result, &operatorbasepb.DomainModel{
			DomainName: src.DomainName,
			DomainType: src.DomainType,
		})
	}
	return result
}
