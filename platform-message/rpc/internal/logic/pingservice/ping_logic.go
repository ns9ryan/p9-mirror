package pingservicelogic

import (
	"context"

	"oa.98ent.com/p9/platform-message/rpc/internal/svc"
	"oa.98ent.com/p9/platform-message/rpc/pb/message"

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

func (l *PingLogic) Ping(in *message.PingRequest) (*message.PingResponse, error) {
	return &message.PingResponse{
		Message: "pong",
	}, nil
}
