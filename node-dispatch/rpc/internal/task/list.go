package task

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/dispatchtask"
	"oa.98ent.com/p9/node-dispatch/rpc/ent/node"
)

// ListRequest 调度任务列表请求
type ListRequest struct {
	Page     int64  // 页码
	PageSize int64  // 每页数量
	Keyword  string // 搜索关键字
	TaskType string // 任务类型
	Status   *int64 // 任务状态
	NodeCode string // 执行节点编码
}

// ListItem 调度任务列表项
type ListItem struct {
	Task     *ent.DispatchTask // 调度任务
	NodeCode string            // 执行节点编码
}

// ListResult 调度任务列表结果
type ListResult struct {
	Total int64       // 数据总数
	List  []*ListItem // 调度任务列表
}

// List 查询调度任务列表
func (s *Service) List(ctx context.Context, req ListRequest) (*ListResult, error) {
	// 创建任务查询
	query := s.db.DispatchTask.Query()

	// 搜索任务编号或请求编号
	if req.Keyword != "" {
		query = query.Where(
			dispatchtask.Or(
				dispatchtask.TaskNoContains(req.Keyword),    // 调度中心生成的全局唯一任务编号
				dispatchtask.RequestNoContains(req.Keyword), // 调用方生成的请求编号
			),
		)
	}

	// 按任务类型筛选
	if req.TaskType != "" {
		query = query.Where(dispatchtask.TaskTypeEQ(req.TaskType)) // 任务类型
	}

	// 按任务状态筛选
	if req.Status != nil {
		query = query.Where(dispatchtask.StatusEQ(*req.Status)) // 任务状态: 1待执行, 2执行中, 3成功, 4失败
	}

	// 按执行节点筛选
	if req.NodeCode != "" {
		query = query.Where(
			dispatchtask.HasNodeWith(
				node.CodeEQ(req.NodeCode),
			),
		)
	}

	// 查询数据总数
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("统计任务数量失败: %w", err)
	}

	// 查询当前页任务及执行节点
	taskList, err := query.
		WithNode().
		Order(ent.Desc(dispatchtask.FieldCreatedAt), ent.Desc(dispatchtask.FieldID)).
		Limit(int(req.PageSize)).
		Offset(int((req.Page - 1) * req.PageSize)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}

	// 组装列表结果
	list := make([]*ListItem, 0, len(taskList))
	for _, taskData := range taskList {
		// 获取执行节点
		nodeData, err := taskData.Edges.NodeOrErr()
		if err != nil {
			return nil, fmt.Errorf("获取任务执行节点失败: %w", err)
		}

		list = append(list, &ListItem{
			Task:     taskData,      // 调度任务
			NodeCode: nodeData.Code, // 执行节点编码
		})
	}

	return &ListResult{
		Total: int64(total), // 数据总数
		List:  list,         // 调度任务列表
	}, nil
}
