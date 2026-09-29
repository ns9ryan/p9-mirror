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

func (l *ListDispatchTasksLogic) ListDispatchTasks(req *types.ListDispatchTasksRequest) (resp *types.ListDispatchTasksResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
