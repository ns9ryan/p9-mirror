package operatornodeservicelogic

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/rpc/internal/svc"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveLogic {
	return &SaveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存分站部署节点
func (l *SaveLogic) Save(in *operatornodepb.SaveOperatorNodeRequest) (*operatornodepb.SaveOperatorNodeResponse, error) {
	// todo: add your logic here and delete this line

	return &operatornodepb.SaveOperatorNodeResponse{}, nil
}
