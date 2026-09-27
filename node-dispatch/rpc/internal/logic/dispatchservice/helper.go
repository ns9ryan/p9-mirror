package dispatchservicelogic

import (
	"oa.98ent.com/p9/node-dispatch/rpc/ent"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
)

// toTaskInfo 转换调度任务信息
func toTaskInfo(data *ent.DispatchTask, nodeCode string) *dispatchpb.TaskInfo {
	taskInfo := &dispatchpb.TaskInfo{
		TaskNo:    data.TaskNo,                // 调度任务编号
		RequestNo: data.RequestNo,             // 调用方请求编号
		Target:    data.Target,                // 目标服务
		TaskType:  data.TaskType,              // 任务类型
		Params:    string(data.Params),        // 任务参数
		NodeCode:  nodeCode,                   // 执行节点编码
		Status:    data.Status,                // 任务状态: 1待执行, 2执行中, 3成功, 4失败
		CreatedAt: data.CreatedAt.UnixMilli(), // 创建时间, Unix毫秒时间戳
		UpdatedAt: data.UpdatedAt.UnixMilli(), // 更新时间, Unix毫秒时间戳
	}

	// 设置执行结果
	if len(data.Result) > 0 {
		taskInfo.Result = new(string(data.Result))
	}

	// 设置失败原因
	if data.ErrorMessage != nil {
		taskInfo.ErrorMessage = new(*data.ErrorMessage)
	}

	// 设置开始执行时间
	if data.StartedAt != nil {
		taskInfo.StartedAt = new(data.StartedAt.UnixMilli())
	}

	// 设置执行结束时间
	if data.FinishedAt != nil {
		taskInfo.FinishedAt = new(data.FinishedAt.UnixMilli())
	}

	return taskInfo
}
