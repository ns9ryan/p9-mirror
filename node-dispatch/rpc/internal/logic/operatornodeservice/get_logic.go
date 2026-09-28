package operatornodeservicelogic

import (
	"context"

	"oa.98ent.com/p9/node-dispatch/rpc/internal/svc"
	"oa.98ent.com/p9/node-dispatch/rpc/pb/nodedispatchrpc/operatornodepb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取分站部署节点
func (l *GetLogic) Get(in *operatornodepb.GetOperatorNodeRequest) (*operatornodepb.GetOperatorNodeResponse, error) {
	// todo: add your logic here and delete this line

	return &operatornodepb.GetOperatorNodeResponse{}, nil
}
