package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"oa.98ent.com/p9/node-dispatch/rpc/ent/dispatchtask"
)

// MarkResultRequest 标记任务执行结果请求
type MarkResultRequest struct {
	TaskNo       string          // 任务编号
	NodeCode     string          // 节点编码
	Success      bool            // 是否执行成功
	Result       json.RawMessage // 执行结果
	ErrorMessage string          // 失败原因
}

// MarkResult 标记任务执行结果
func (s *Service) MarkResult(ctx context.Context, req MarkResultRequest) error {
	// 查询调度任务及执行节点
	taskData, err := s.db.DispatchTask.
		Query().
		Where(dispatchtask.TaskNoEQ(req.TaskNo)).
		WithNode().
		Only(ctx)
	if err != nil {
		return fmt.Errorf("查询调度任务失败: %w", err)
	}

	// 获取执行节点
	nodeData, err := taskData.Edges.NodeOrErr()
	if err != nil {
		return fmt.Errorf("获取任务执行节点失败: %w", err)
	}

	// 校验执行节点
	if nodeData.Code != req.NodeCode {
		return fmt.Errorf("任务执行节点不匹配: expect=%s actual=%s", nodeData.Code, req.NodeCode)
	}

	// 已完成的任务不重复处理
	if taskData.Status == StatusSuccess || taskData.Status == StatusFailed {
		return nil
	}

	// 创建任务结果更新
	taskUpdate := s.db.DispatchTask.
		Update().
		Where(
			dispatchtask.IDEQ(taskData.ID),                      // 任务ID
			dispatchtask.StatusIn(StatusPending, StatusRunning), // 任务状态: 1待执行, 2执行中
		).
		SetFinishedAt(time.Now())

	if req.Success {
		// 标记任务执行成功
		taskUpdate.
			SetStatus(StatusSuccess). // 任务状态: 3成功
			SetResult(req.Result)     // 任务执行结果
	} else {
		// 标记任务执行失败
		taskUpdate.
			SetStatus(StatusFailed).          // 任务状态: 4失败
			SetErrorMessage(req.ErrorMessage) // 任务执行失败原因
	}

	// 保存任务执行结果
	affected, err := taskUpdate.Save(ctx)
	if err != nil {
		return fmt.Errorf("更新任务执行结果失败: %w", err)
	}

	// 任务已经被其他结果处理时忽略本次结果
	if affected == 0 {
		return nil
	}

	return nil
}
