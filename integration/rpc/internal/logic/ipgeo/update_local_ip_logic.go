package ipgeo

import (
	"context"

	geo "oa.98ent.com/p9/integration/rpc/internal/ipgeo"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateLocalIpLogic 更新本地 IP（RPC updateLocalIp）。
type UpdateLocalIpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLocalIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLocalIpLogic {
	return &UpdateLocalIpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateLocalIp 按主键更新本地 IP，并失效新旧 IP 缓存。
func (l *UpdateLocalIpLogic) UpdateLocalIp(in *integration.LocalIP) (*integration.LocalIP, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, status.Error(codes.Unavailable, "ip geo not ready")
	}
	if in.GetId() <= 0 {
		return nil, mapStoreErr(geo.ErrBadProvider)
	}
	row, err := l.svcCtx.Deps.IpGeo.UpdateLocalIP(l.ctx, localFromPB(in))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	return localToPB(row), nil
}
