package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	operatorgamepb "oa.98ent.com/p9/operator-game/rpc/pb/operator_game"
)

// publishGameAllocationParams 发布游戏资源分配任务参数
type publishGameAllocationParams struct {
	OperatorCode string `json:"operator_code"` // 分站全局唯一业务编码
}

// executePublishGameAllocation 执行发布游戏资源分配任务
func (s *Service) executePublishGameAllocation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)
	logger.Infof("开始执行发布游戏资源分配任务: params=%s", string(params))

	// 解析任务参数
	var taskParams publishGameAllocationParams
	if err := json.Unmarshal(params, &taskParams); err != nil {
		logger.Errorf("解析发布游戏资源分配任务参数失败: %v", err)
		return nil, fmt.Errorf("解析任务参数失败: %w", err)
	}

	// 整理并校验分站业务编码
	operatorCode := strings.TrimSpace(taskParams.OperatorCode)
	if operatorCode == "" {
		logger.Error("operator_code不能为空")
		return nil, fmt.Errorf("operator_code不能为空")
	}

	logger.Infow("开始发布游戏资源分配", logx.Field("operator_code", operatorCode))

	// 调用分站游戏RPC服务同步已发布的游戏数据
	// operator-game RPC会调用platform-game RPC获取总网资源数据，并插入到分站数据库
	rpcResp, err := s.operatorGameRpc.SyncPublishedData(ctx, &operatorgamepb.SyncPublishedDataRequest{
		OpCode: operatorCode,
	})
	if err != nil {
		logger.Errorf("调用分站游戏RPC同步已发布数据失败: operator_code=%s, error=%v", operatorCode, err)
		return nil, fmt.Errorf("同步已发布数据失败: %w", err)
	}

	// 检查同步结果
	if rpcResp == nil {
		logger.Errorf("分站游戏RPC返回结果为空: operator_code=%s", operatorCode)
		return nil, fmt.Errorf("RPC返回结果为空")
	}

	if !rpcResp.GetSuccess() {
		logger.Errorf("游戏资源分配发布失败",
			logx.Field("operator_code", operatorCode),
			logx.Field("message", rpcResp.GetMessage()),
		)
		return nil, fmt.Errorf("游戏资源分配发布失败: %s", rpcResp.GetMessage())
	}

	// 记录同步统计信息
	logger.Infow("游戏资源分配发布成功",
		logx.Field("operator_code", operatorCode),
		logx.Field("category_success", rpcResp.GetCategoryStat().GetSuccess()),
		logx.Field("provider_success", rpcResp.GetProviderStat().GetSuccess()),
		logx.Field("channel_success", rpcResp.GetChannelStat().GetSuccess()),
		logx.Field("game_success", rpcResp.GetGameStat().GetSuccess()),
	)

	// 当前任务无需返回业务结果
	return nil, nil
}
