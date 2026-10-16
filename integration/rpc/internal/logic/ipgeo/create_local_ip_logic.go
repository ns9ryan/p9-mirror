package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateLocalIpLogic 新增本地 IP（RPC createLocalIp）。
type CreateLocalIpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLocalIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLocalIpLogic {
	return &CreateLocalIpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateLocalIp 写入 ip_geo_local，并失效该 IP 查询缓存。
func (l *CreateLocalIpLogic) CreateLocalIp(in *integration.LocalIP) (*integration.LocalIP, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, status.Error(codes.Unavailable, "ip geo not ready")
	}
	row, err := l.svcCtx.Deps.IpGeo.CreateLocalIP(l.ctx, localFromPB(in))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	return localToPB(row), nil
}
