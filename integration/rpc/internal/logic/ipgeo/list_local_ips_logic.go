package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListLocalIpsLogic 本地 IP 分页列表（RPC listLocalIps）。
type ListLocalIpsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLocalIpsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLocalIpsLogic {
	return &ListLocalIpsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListLocalIps 分页列出本地 IP；ip 非空时做包含匹配。
func (l *ListLocalIpsLogic) ListLocalIps(in *integration.LocalIPListReq) (*integration.LocalIPListResp, error) {
	if l.svcCtx.Deps == nil || l.svcCtx.Deps.IpGeo == nil {
		return nil, status.Error(codes.Unavailable, "ip geo not ready")
	}
	page, err := l.svcCtx.Deps.IpGeo.ListLocalIPs(l.ctx, int(in.GetPage()), int(in.GetPageSize()), in.GetIp())
	if err != nil {
		return nil, mapStoreErr(err)
	}
	out := make([]*integration.LocalIP, 0, len(page.List))
	for i := range page.List {
		out = append(out, localToPB(&page.List[i]))
	}
	return &integration.LocalIPListResp{List: out, Total: int64(page.Total)}, nil
}
