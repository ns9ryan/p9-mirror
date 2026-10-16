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

// publishBasicAllocationParams 发布基础资源分配任务参数
type publishBasicAllocationParams struct {
	OperatorCode string `json:"operator_code"` // 分站全局唯一业务编码
}

// executePublishBasicAllocation 执行发布基础资源分配任务
func (s *Service) executePublishBasicAllocation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)
	logger.Infof("开始执行发布基础资源分配任务: params=%s", string(params))

	// 解析任务参数
	var taskParams publishBasicAllocationParams
	if err := json.Unmarshal(params, &taskParams); err != nil {
		logger.Errorf("解析发布基础资源分配任务参数失败: %v", err)
		return nil, fmt.Errorf("解析任务参数失败: %w", err)
	}

	// 整理并校验分站业务编码
	operatorCode := strings.TrimSpace(taskParams.OperatorCode)
	if operatorCode == "" {
		logger.Error("operator_code不能为空")
		return nil, fmt.Errorf("operator_code不能为空")
	}

	logger.Infow("开始发布基础资源分配", logx.Field("operator_code", operatorCode))

	// ==================== 从总网获取分站初始化数据 ====================

	// 调用总网分站RPC获取分站初始化数据
	// 包括：语言分配、经营地区分配、代理子线路分配
	logger.Debugw(
		"准备调用platformOperatorRpc.GetInitializationData",
		logx.Field("operator_code", operatorCode),
	)

	// 添加上下文超时信息
	deadline, hasDeadline := ctx.Deadline()
	logger.Infow(
		"RPC调用上下文信息(GetInitializationData)",
		logx.Field("operator_code", operatorCode),
		logx.Field("has_deadline", hasDeadline),
		logx.Field("deadline", fmt.Sprintf("%v", deadline)),
	)

	initDataResp, err := s.platformOperatorRpc.GetInitializationData(ctx, &platformoperatorpb.GetInitializationDataRequest{
		OperatorCode: operatorCode,
	})
	if err != nil {
		logger.Errorw(
			"调用总网RPC获取分站初始化数据失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error_type", fmt.Sprintf("%T", err)),
			logx.Field("error_message", err.Error()),
			logx.Field("error_full", fmt.Sprintf("%+v", err)),
		)
		return nil, fmt.Errorf("获取分站初始化数据失败: %w", err)
	}

	// 检查初始化数据是否完整
	if initDataResp == nil {
		logger.Errorf("分站初始化数据为空: operator_code=%s", operatorCode)
		return nil, fmt.Errorf("分站初始化数据为空")
	}

	// ==================== 初始化分站基础资源 ====================

	logger.Infow("开始初始化分站基础资源",
		logx.Field("operator_code", operatorCode),
		logx.Field("language_count", len(initDataResp.LanguageCodes)),
		logx.Field("region_count", len(initDataResp.RegionCodes)),
		logx.Field("agent_line_count", len(initDataResp.AgentLineCodes)),
	)

	// 调用分站基础RPC初始化分站基础资源
	// operator-base RPC会根据提供的语言、地区、代理线路编码初始化分站数据库
	logger.Debugw(
		"准备调用operatorBaseRpc.InitializeBasicResources",
		logx.Field("operator_code", operatorCode),
		logx.Field("language_codes", fmt.Sprintf("%v", initDataResp.LanguageCodes)),
		logx.Field("region_codes", fmt.Sprintf("%v", initDataResp.RegionCodes)),
		logx.Field("agent_line_codes", fmt.Sprintf("%v", initDataResp.AgentLineCodes)),
	)

	// 添加上下文超时信息
	deadline2, hasDeadline2 := ctx.Deadline()
	logger.Infow(
		"RPC调用上下文信息(InitializeBasicResources)",
		logx.Field("operator_code", operatorCode),
		logx.Field("has_deadline", hasDeadline2),
		logx.Field("deadline", fmt.Sprintf("%v", deadline2)),
	)

	logger.Infof("开始RPC调用: operatorBaseRpc.InitializeBasicResources(operator_code=%s)", operatorCode)

	// 添加网络连接诊断信息
	logger.Infow(
		"网络诊断: 准备调用operator-base-rpc服务",
		logx.Field("operator_code", operatorCode),
		logx.Field("service", "operator-base"),
		logx.Field("method", "InitializeBasicResources"),
	)

	rpcResp, err := s.operatorBaseRpc.InitializeBasicResources(ctx, &operatorbasepb.InitializeBasicResourcesRequest{
		OperatorCode:   operatorCode,
		LanguageCodes:  initDataResp.LanguageCodes,
		RegionCodes:    initDataResp.RegionCodes,
		AgentLineCodes: initDataResp.AgentLineCodes,
	})
	if err != nil {
		logger.Errorw(
			"调用分站基础RPC初始化基础资源失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("error_type", fmt.Sprintf("%T", err)),
			logx.Field("error_message", err.Error()),
			logx.Field("error_full", fmt.Sprintf("%+v", err)),
		)
		return nil, fmt.Errorf("初始化基础资源失败: %w", err)
	}

	// 检查初始化结果
	if rpcResp == nil {
		logger.Errorf("分站基础RPC返回结果为空: operator_code=%s", operatorCode)
		return nil, fmt.Errorf("RPC返回结果为空")
	}

	if !rpcResp.GetSuccess() {
		logger.Errorf("分站基础资源初始化失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("message", rpcResp.GetMessage()),
		)
		return nil, fmt.Errorf("分站基础资源初始化失败: %s", rpcResp.GetMessage())
	}

	// 记录初始化统计信息
	logger.Infow("分站基础资源初始化成功",
		logx.Field("operator_code", operatorCode),
		logx.Field("language_success", rpcResp.GetLanguageStat().GetSuccess()),
		logx.Field("region_success", rpcResp.GetRegionStat().GetSuccess()),
		logx.Field("agent_line_success", rpcResp.GetAgentLineStat().GetSuccess()),
	)

	// 当前任务无需返回业务结果
	return nil, nil
}
