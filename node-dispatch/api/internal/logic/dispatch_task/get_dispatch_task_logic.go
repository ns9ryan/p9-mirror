// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dispatch_task

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/dispatchpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDispatchTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDispatchTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDispatchTaskLogic {
	return &GetDispatchTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetDispatchTask 获取调度任务
func (l *GetDispatchTaskLogic) GetDispatchTask(req *types.GetDispatchTaskRequest) (resp *types.GetDispatchTaskResponse, err error) {
	// 获取调度任务
	result, err := l.svcCtx.DispatchRpc.GetTask(
		l.ctx,
		&dispatchpb.GetTaskRequest{
			TaskNo: &req.TaskNo, // 调度任务编号
		},
	)
	if err != nil {
		return nil, err
	}

	// 返回调度任务信息
	return &types.GetDispatchTaskResponse{
		DispatchTaskInfo: toDispatchTaskInfo(result.Task), // 调度任务信息
	}, nil
}
