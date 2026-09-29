// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dispatch_task

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/api/internal/svc"
	"oa.98ent.com/p9/node-dispatch/api/internal/types"

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

func (l *GetDispatchTaskLogic) GetDispatchTask(req *types.GetDispatchTaskRequest) (resp *types.GetDispatchTaskResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
