package operatorservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-operator/rpc/internal/svc"
	"oa.98ent.com/p9/platform-operator/rpc/pb/platformoperatorrpc/operatorpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandlePublishResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandlePublishResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandlePublishResultLogic {
	return &HandlePublishResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理分站发布结果
func (l *HandlePublishResultLogic) HandlePublishResult(in *operatorpb.HandlePublishResultRequest) (*operatorpb.HandlePublishResultResponse, error) {
	// todo: add your logic here and delete this line

	return &operatorpb.HandlePublishResultResponse{}, nil
}
