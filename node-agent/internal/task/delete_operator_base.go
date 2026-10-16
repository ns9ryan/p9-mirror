package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	operatorpb "oa.98ent.com/p9/operator-base/rpc/pb/operatorbaserpc/operatorpb"
)

// deleteOperatorBaseParams 删除operator-base中的分站数据参数
type deleteOperatorBaseParams struct {
	OperatorCode string `json:"operator_code"`
}

// executeDeleteOperatorBase 执行删除operator-base中的分站数据
func (s *Service) executeDeleteOperatorBase(ctx context.Context, rawParams json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)

	// ==================== 第一步：解析参数 ====================

	var params deleteOperatorBaseParams
	if err := json.Unmarshal(rawParams, &params); err != nil {
		logger.Errorf("解析删除参数失败: error=%v", err)
		return nil, fmt.Errorf("解析删除参数失败: %w", err)
	}

	operatorCode := params.OperatorCode
	if operatorCode == "" {
		logger.Errorf("分站编码为空")
		return nil, fmt.Errorf("分站编码为空")
	}

	logger.Infof("开始执行删除operator-base分站数据: operator_code=%s", operatorCode)

	// ==================== 第二步：调用 operator-base 删除分站基础资源数据 ====================

	deleteResp, err := s.operatorBaseRpc.DeleteOperatorBasicResources(ctx, &operatorpb.DeleteOperatorBasicResourcesRequest{
		OperatorCode: operatorCode,
	})
	if err != nil {
		logger.Errorf("调用operator-base.DeleteOperatorBasicResources失败: operator_code=%s, error=%v", operatorCode, err)
		return nil, fmt.Errorf("删除operator-base分站基础资源数据失败: %w", err)
	}

	if deleteResp == nil {
		logger.Errorf("operator-base返回删除响应为空: operator_code=%s", operatorCode)
		return nil, fmt.Errorf("operator-base返回删除响应为空")
	}

	if !deleteResp.Success {
		logger.Errorf("删除operator-base分站基础资源数据失败: operator_code=%s, message=%s", operatorCode, deleteResp.Message)
		return nil, fmt.Errorf("删除operator-base分站基础资源数据失败: %s", deleteResp.Message)
	}

	logger.Infof("删除operator-base分站基础资源数据成功: operator_code=%s, 删除统计: 域名=%d, 语言=%d, 地区=%d, 代理线=%d",
		operatorCode,
		deleteResp.DomainDeleted,
		deleteResp.LanguageDeleted,
		deleteResp.RegionDeleted,
		deleteResp.AgentLineDeleted,
	)

	return nil, nil
}
