// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dispatch_task

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDispatchTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDispatchTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDispatchTasksLogic {
	return &ListDispatchTasksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListDispatchTasks 获取调度任务列表
func (l *ListDispatchTasksLogic) ListDispatchTasks(req *types.ListDispatchTasksRequest) (resp *types.ListDispatchTasksResponse, err error) {
	// 获取调度任务列表
	result, err := l.svcCtx.DispatchRpc.ListTask(
		l.ctx,
		&dispatchpb.ListTasksRequest{
			Page:     req.Page,     // 页码, 从1开始
			PageSize: req.PageSize, // 每页数量
			Keyword:  req.Keyword,  // 搜索关键字, 匹配任务编号或请求编号
			TaskType: req.TaskType, // 任务类型
			Status:   req.Status,   // 任务状态: 1待执行, 2执行中, 3成功, 4失败
			NodeCode: req.NodeCode, // 执行节点编码
		},
	)
	if err != nil {
		return nil, err
	}

	// 转换调度任务列表
	list := make([]types.DispatchTaskInfo, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toDispatchTaskInfo(item))
	}

	// 返回调度任务列表
	return &types.ListDispatchTasksResponse{
		Total: result.Total, // 数据总数
		List:  list,         // 调度任务列表
	}, nil
}
