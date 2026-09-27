package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/node"
	"oa.98ent.com/p9/node-dispatch/rpc/internal/connection"
	"oa.98ent.com/p9/node-dispatch/rpc/internal/protocol"
)

// SubmitRequest 提交调度任务请求
type SubmitRequest struct {
	RequestNo string          // 调用方请求编号
	Target    string          // 目标服务
	TaskType  string          // 任务类型
	NodeCode  string          // 执行节点编码
	Params    json.RawMessage // 任务参数
}

// SubmitResponse 提交调度任务响应
type SubmitResponse struct {
	TaskNo string // 调度任务编号
}

// Submit 提交调度任务
func (s *Service) Submit(ctx context.Context, req SubmitRequest) (*SubmitResponse, error) {
	// 检查任务幂等
	taskData, exists, err := s.checkIdempotent(ctx, req)
	if err != nil {
		return nil, err
	}
	if exists {
		return &SubmitResponse{
			TaskNo: taskData.TaskNo, // 调度任务编号
		}, nil
	}

	// 获取执行节点
	nodeData, err := s.db.Node.
		Query().
		Where(node.CodeEQ(req.NodeCode)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		return nil, fmt.Errorf("查询执行节点失败: %w", err)
	}

	// 检查节点状态
	if nodeData.Status != 1 {
		return nil, status.Error(codes.FailedPrecondition, "node is disabled")
	}

	// 检查节点连接状态
	if !s.connections.IsOnline(req.NodeCode) {
		return nil, status.Error(codes.Unavailable, "node is offline")
	}

	// 创建调度任务
	taskData, err = s.createTask(ctx, req, nodeData.ID)
	if err != nil {
		// 并发提交相同request_no时, 由数据库唯一约束完成最终幂等保护
		if ent.IsConstraintError(err) {
			existingTask, existing, idempotentErr := s.checkIdempotent(ctx, req)
			if idempotentErr != nil {
				return nil, idempotentErr
			}
			if existing {
				return &SubmitResponse{
					TaskNo: existingTask.TaskNo, // 已存在的调度任务编号
				}, nil
			}
		}

		return nil, err
	}

	// 下发任务
	if err = s.dispatch(ctx, taskData.TaskNo, req); err != nil {
		// 标记任务下发失败
		s.markDispatchFailed(ctx, taskData.ID, err.Error())

		// 节点在任务创建后断开连接
		if errors.Is(err, connection.ErrNodeOffline) {
			return nil, status.Error(codes.Unavailable, "node is offline")
		}

		return nil, status.Error(codes.Unavailable, "failed to dispatch task")
	}

	// 返回任务编号
	return &SubmitResponse{
		TaskNo: taskData.TaskNo, // 调度任务编号
	}, nil
}

// createTask 创建调度任务
func (s *Service) createTask(ctx context.Context, req SubmitRequest, nodeID int64) (*ent.DispatchTask, error) {
	taskData, err := s.db.DispatchTask.
		Create().
		SetTaskNo(uuid.NewString()). // 调度任务编号
		SetRequestNo(req.RequestNo). // 调用方请求编号
		SetTarget(req.Target).       // 目标服务
		SetTaskType(req.TaskType).   // 任务类型
		SetParams(req.Params).       // 任务参数
		SetNodeID(nodeID).           // 执行节点ID
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建调度任务失败: %w", err)
	}

	return taskData, nil
}

// dispatch 下发调度任务
func (s *Service) dispatch(ctx context.Context, taskNo string, req SubmitRequest) error {
	// 编码任务数据
	data, err := json.Marshal(protocol.TaskDispatchData{
		TaskNo:   taskNo,       // 调度任务编号
		Target:   req.Target,   // 目标服务
		TaskType: req.TaskType, // 任务类型
		Params:   req.Params,   // 任务参数
	})
	if err != nil {
		return fmt.Errorf("编码任务数据失败: %w", err)
	}

	// 发送任务消息
	if err = s.connections.Send(ctx, req.NodeCode, protocol.Message{
		Type: protocol.MessageTypeTaskDispatch, // 消息类型
		Data: data,                             // 消息数据
	}); err != nil {
		return fmt.Errorf("下发任务失败: %w", err)
	}

	return nil
}
