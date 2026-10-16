package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteLocalIpLogic 删除本地 IP（RPC deleteLocalIp）。
type DeleteLocalIpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteLocalIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLocalIpLogic {
	return &DeleteLocalIpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteLocalIp 按主键删除本地 IP，并失效缓存。
func (l *DeleteLocalIpLogic) DeleteLocalIp(in *integration.IDReq) (*integration.Empty, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, status.Error(codes.Unavailable, "ip geo not ready")
	}
	if err := l.svcCtx.Deps.IpGeo.DeleteLocalIP(l.ctx, in.GetId()); err != nil {
		return nil, mapStoreErr(err)
	}
	return &integration.Empty{}, nil
}
