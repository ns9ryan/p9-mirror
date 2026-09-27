package task

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/dispatchtask"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetRequest 获取调度任务请求
type GetRequest struct {
	TaskNo    string // 调度任务编号
	RequestNo string // 调用方请求编号
}

// GetResult 获取调度任务结果
type GetResult struct {
	Task     *ent.DispatchTask // 调度任务
	NodeCode string            // 执行节点编码
}

// Get 获取调度任务
func (s *Service) Get(ctx context.Context, req GetRequest) (*GetResult, error) {
	// 创建任务查询
	query := s.db.DispatchTask.Query()

	// 根据指定编号查询任务
	if req.TaskNo != "" {
		query = query.Where(dispatchtask.TaskNoEQ(req.TaskNo)) // 调度中心生成的全局唯一任务编号
	} else {
		query = query.Where(dispatchtask.RequestNoEQ(req.RequestNo)) // 调用方生成的请求编号
	}

	// 获取调度任务及执行节点
	taskData, err := query.
		WithNode().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		return nil, fmt.Errorf("查询调度任务失败: %w", err)
	}

	// 获取执行节点
	nodeData, err := taskData.Edges.NodeOrErr()
	if err != nil {
		return nil, fmt.Errorf("获取任务执行节点失败: %w", err)
	}

	// 返回任务详情
	return &GetResult{
		Task:     taskData,      // 调度任务
		NodeCode: nodeData.Code, // 执行节点编码
	}, nil
}
