// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package dispatch_callback

import (
	"context"

	"oa.98ent.com/p9/platform-operator/api/internal/svc"
	"oa.98ent.com/p9/platform-operator/api/internal/types"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"

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

// HandleDispatchTaskResult 处理调度任务结果回调
func (l *HandleDispatchTaskResultLogic) HandleDispatchTaskResult(req *types.HandleDispatchTaskResultRequest) (resp *types.HandleDispatchTaskResultResponse, err error) {
	// 根据调度任务编号同步分站发布结果
	_, err = l.svcCtx.OperatorRpc.HandlePublishResult(l.ctx, &operatorpb.HandlePublishResultRequest{
		TaskNo: req.TaskNo, // 调度任务编号
	})
	if err != nil {
		return nil, err
	}

	// 返回回调处理结果
	return &types.HandleDispatchTaskResultResponse{}, nil
}
