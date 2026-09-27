package dispatchservicelogic

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"oa.98ent.com/p9/node-dispatch/rpc/internal/svc"
	"oa.98ent.com/p9/node-dispatch/rpc/internal/task"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"
)

type GetTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskLogic {
	return &GetTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTask 获取调度任务
func (l *GetTaskLogic) GetTask(in *dispatchpb.GetTaskRequest) (*dispatchpb.GetTaskResponse, error) {
	// 整理查询参数
	taskNo := strings.TrimSpace(in.GetTaskNo())
	requestNo := strings.TrimSpace(in.GetRequestNo())

	// 校验查询参数
	if taskNo == "" && requestNo == "" {
		return nil, status.Error(codes.InvalidArgument, "task_no or request_no is required")
	}
	if taskNo != "" && requestNo != "" {
		return nil, status.Error(codes.InvalidArgument, "task_no and request_no cannot be provided together")
	}

	// 查询调度任务
	result, err := l.svcCtx.Task.Get(l.ctx, task.GetRequest{
		TaskNo:    taskNo,    // 调度任务编号
		RequestNo: requestNo, // 调用方请求编号
	})
	if err != nil {
		l.Logger.Errorw(
			"获取调度任务失败",
			logx.Field("task_no", taskNo),
			logx.Field("request_no", requestNo),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	// 校验查询结果
	if result == nil || result.Task == nil {
		l.Logger.Errorw(
			"获取调度任务结果为空",
			logx.Field("task_no", taskNo),
			logx.Field("request_no", requestNo),
		)
		return nil, status.Error(codes.Internal, "task result is empty")
	}

	// 返回调度任务
	return &dispatchpb.GetTaskResponse{
		Task: toTaskInfo(result.Task, result.NodeCode), // 调度任务信息
	}, nil
}
