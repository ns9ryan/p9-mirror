package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-game/rpc/pb/operator_game"
)

// deleteOperatorGameParams 删除operator-game中的分站游戏数据参数
type deleteOperatorGameParams struct {
	OperatorCode string `json:"operator_code"`
}

// executeDeleteOperatorGame 执行删除operator-game中的分站游戏数据
func (s *Service) executeDeleteOperatorGame(ctx context.Context, rawParams json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)

	// ==================== 第一步：解析参数 ====================

	var params deleteOperatorGameParams
	if err := json.Unmarshal(rawParams, &params); err != nil {
		logger.Errorf("解析删除参数失败: error=%v", err)
		return nil, fmt.Errorf("解析删除参数失败: %w", err)
	}

	operatorCode := params.OperatorCode
	if operatorCode == "" {
		logger.Errorf("分站编码为空")
		return nil, fmt.Errorf("分站编码为空")
	}

	logger.Infof("开始执行删除operator-game分站游戏数据: operator_code=%s", operatorCode)

	// ==================== 第二步：调用 operator-game 删除分站游戏数据 ====================

	delResp, err := s.operatorGameRpc.DeleteOperatorData(ctx, &operator_game.DeleteOperatorDataRequest{
		OpCode: operatorCode,
	})
	if err != nil {
		logger.Errorf("调用operator-game.DeleteOperatorData失败: operator_code=%s, error=%v", operatorCode, err)
		return nil, fmt.Errorf("删除operator-game分站游戏数据失败: %w", err)
	}

	if delResp == nil {
		logger.Errorf("operator-game返回删除响应为空: operator_code=%s", operatorCode)
		return nil, fmt.Errorf("operator-game返回删除响应为空")
	}

	if !delResp.Success {
		logger.Errorf("删除operator-game分站游戏数据失败: operator_code=%s, message=%s", operatorCode, delResp.Message)
		return nil, fmt.Errorf("删除operator-game分站游戏数据失败: %s", delResp.Message)
	}

	logger.Infof("删除operator-game分站游戏数据成功: operator_code=%s, 删除统计: 分类=%d, 服务商=%d, 渠道=%d, 游戏=%d",
		operatorCode,
		delResp.CategoryDeleted,
		delResp.ProviderDeleted,
		delResp.ChannelDeleted,
		delResp.GameDeleted,
	)

	return nil, nil
}
