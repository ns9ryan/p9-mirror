// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dispatch_callback

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleDispatchTaskResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleDispatchTaskResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleDispatchTaskResultLogic {
	return &HandleDispatchTaskResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleDispatchTaskResultLogic) HandleDispatchTaskResult(req *types.HandleDispatchTaskResultRequest) (resp *types.HandleDispatchTaskResultResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
