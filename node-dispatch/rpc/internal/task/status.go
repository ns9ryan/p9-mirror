package task

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/node-dispatch/rpc/ent/dispatchtask"
)

const (
	StatusPending int64 = 1 // 待执行
	StatusRunning int64 = 2 // 执行中
	StatusSuccess int64 = 3 // 成功
	StatusFailed  int64 = 4 // 失败
)

// MarkRunning 标记任务执行中
func (s *Service) MarkRunning(ctx context.Context, taskNo string, nodeCode string) error {
	// 查询调度任务及执行节点
	taskData, err := s.db.DispatchTask.
		Query().
		Where(dispatchtask.TaskNoEQ(taskNo)).
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
	if nodeData.Code != nodeCode {
		return fmt.Errorf("任务执行节点不匹配: expect=%s actual=%s", nodeData.Code, nodeCode)
	}

	// 已经开始或结束的任务不重复处理
	if taskData.Status != StatusPending {
		return nil
	}

	// 将待执行任务标记为执行中
	affected, err := s.db.DispatchTask.
		Update().
		Where(
			dispatchtask.IDEQ(taskData.ID),
			dispatchtask.StatusEQ(StatusPending), // 任务状态: 1待执行
		).
		SetStatus(StatusRunning). // 任务状态: 2执行中
		SetStartedAt(time.Now()). // 开始执行时间
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新任务执行状态失败: %w", err)
	}

	// 任务已经被其他消息更新时忽略本次ACK
	if affected == 0 {
		return nil
	}

	return nil
}
