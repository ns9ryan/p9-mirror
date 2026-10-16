package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetLocalIpLogic 本地 IP 详情（RPC getLocalIp）。
type GetLocalIpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLocalIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLocalIpLogic {
	return &GetLocalIpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLocalIp 按主键查询本地 IP。
func (l *GetLocalIpLogic) GetLocalIp(in *integration.IDReq) (*integration.LocalIP, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, status.Error(codes.Unavailable, "ip geo not ready")
	}
	row, err := l.svcCtx.Deps.IpGeo.GetLocalIP(l.ctx, in.GetId())
	if err != nil {
		return nil, mapStoreErr(err)
	}
	return localToPB(row), nil
}
