package pingservicelogic

import (
	"context"

	"oa.98ent.com/p9/operator-crm/service/settlement/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/settlement/rpc/pb/settlementrpc/pingpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 健康检查
func (l *PingLogic) Ping(in *pingpb.PingRequest) (*pingpb.PingResponse, error) {
	// todo: add your logic here and delete this line

	return &pingpb.PingResponse{}, nil
}
