package dispatch_task

import (
	"oa.98ent.com/p9/node-dispatch/api/internal/types"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
)

// toDispatchTaskInfo 转换调度任务信息
func toDispatchTaskInfo(data *dispatchpb.TaskInfo) types.DispatchTaskInfo {
	return types.DispatchTaskInfo{
		TaskNo:       data.TaskNo,       // 调度任务编号
		RequestNo:    data.RequestNo,    // 调用方请求编号
		Target:       data.Target,       // 任务目标服务
		TaskType:     data.TaskType,     // 任务类型
		Params:       data.Params,       // 任务参数, JSON数据
		NodeCode:     data.NodeCode,     // 执行节点编码
		Status:       data.Status,       // 任务状态: 1待执行, 2执行中, 3成功, 4失败
		Result:       data.Result,       // 任务执行结果, JSON数据
		ErrorMessage: data.ErrorMessage, // 任务执行失败原因
		StartedAt:    data.StartedAt,    // 开始执行时间, Unix毫秒时间戳
		FinishedAt:   data.FinishedAt,   // 执行结束时间, Unix毫秒时间戳
		CreatedAt:    data.CreatedAt,    // 创建时间, Unix毫秒时间戳
		UpdatedAt:    data.UpdatedAt,    // 更新时间, Unix毫秒时间戳
	}
}
