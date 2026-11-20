package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	TargetOperatorBase               = "operator-base"                   // 任务目标: 分站基础服务
	TargetOperatorGame               = "operator-game"                   // 任务目标: 分站游戏服务
	TypeCreateOperator               = "CREATE_OPERATOR"                 // 任务类型: 创建分站
	TypePublishBasicAllocation       = "PUBLISH_BASIC_ALLOCATION"        // 任务类型: 发布基础资源分配
	TypePublishDomain                = "PUBLISH_DOMAIN"                  // 任务类型: 发布域名
	TypePublishGameAllocation        = "PUBLISH_GAME_ALLOCATION"         // 任务类型: 发布游戏资源分配
	TypeDeleteOperatorBasicResources = "DELETE_OPERATOR_BASIC_RESOURCES" // 任务类型: 删除分站基础资源数据
	TypeDeleteOperatorDomain         = "DELETE_OPERATOR_DOMAIN"          // 任务类型: 删除分站域名
	TypeDeleteOperatorGameAllocation = "DELETE_OPERATOR_GAME_ALLOCATION" // 任务类型: 删除游戏资源分配
)

// Execute 执行调度任务
func (s *Service) Execute(ctx context.Context, target string, taskType string, params json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)

	// 整理任务信息
	target = strings.TrimSpace(target)
	taskType = strings.TrimSpace(taskType)

	logger.Infow(
		"开始执行调度任务",
		logx.Field("target", target),
		logx.Field("task_type", taskType),
		logx.Field("params_length", len(params)),
	)

	// 根据目标服务执行任务
	switch target {
	case TargetOperatorBase:
		logger.Infof("路由到executeOperatorBaseTask, taskType=%s", taskType)
		return s.executeOperatorBaseTask(ctx, taskType, params)
	case TargetOperatorGame:
		logger.Infof("路由到executeOperatorGameTask, taskType=%s", taskType)
		return s.executeOperatorGameTask(ctx, taskType, params)
	default:
		errMsg := fmt.Sprintf("不支持的任务目标服务: %s", target)
		logger.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
}

// executeOperatorBaseTask 执行分站基础服务任务
func (s *Service) executeOperatorBaseTask(ctx context.Context, taskType string, params json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)
	logger.Infow(
		"executeOperatorBaseTask开始处理",
		logx.Field("task_type", taskType),
	)

	// 根据任务类型执行任务
	switch taskType {
	case TypeCreateOperator:
		logger.Infof("执行CREATE_OPERATOR任务")
		return s.executeCreateOperator(ctx, params)
	case TypePublishBasicAllocation:
		logger.Infof("执行PUBLISH_BASIC_ALLOCATION任务")
		return s.executePublishBasicAllocation(ctx, params)
	case TypePublishDomain:
		logger.Infof("执行PUBLISH_DOMAIN任务")
		return s.executePublishDomain(ctx, params)
	case TypeDeleteOperatorBasicResources:
		logger.Infof("执行DELETE_OPERATOR_BASIC_RESOURCES任务")
		return s.executeDeleteOperatorBase(ctx, params)
	case TypeDeleteOperatorDomain:
		logger.Infof("执行DELETE_OPERATOR_DOMAIN任务")
		return s.executeDeleteOperatorDomain(ctx, params)

	default:
		errMsg := fmt.Sprintf("不支持的operator-base任务类型: %s", taskType)
		logger.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
}

// executeOperatorGameTask 执行分站游戏服务任务
func (s *Service) executeOperatorGameTask(ctx context.Context, taskType string, params json.RawMessage) (json.RawMessage, error) {
	logger := logx.WithContext(ctx)
	logger.Infow(
		"executeOperatorGameTask开始处理",
		logx.Field("task_type", taskType),
	)

	// 根据任务类型执行任务
	switch taskType {
	case TypePublishGameAllocation:
		logger.Infof("执行PUBLISH_GAME_ALLOCATION任务")
		return s.executePublishGameAllocation(ctx, params)
	case TypeDeleteOperatorGameAllocation:
		logger.Infof("执行DELETE_OPERATOR_GAME_ALLOCATION任务")
		return s.executeDeleteOperatorGame(ctx, params)

	default:
		errMsg := fmt.Sprintf("不支持的operator-game任务类型: %s", taskType)
		logger.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
}
