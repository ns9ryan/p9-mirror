package ipgeo

import (
	"context"
	"errors"

	geo "oa.98ent.com/p9/integration/rpc/internal/ipgeo"
	"oa.98ent.com/p9/integration/rpc/internal/svc"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SearchIpLogic IP/GEO 批量查询（RPC searchIp）。
type SearchIpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchIpLogic {
	return &SearchIpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchIp 按租户策略选源查询；租户取自登录态 / metadata，不读请求体 operator_code。
func (l *SearchIpLogic) SearchIp(in *integration.SearchIpReq) (*integration.SearchIpResp, error) {
	if l.svcCtx.IpGeo == nil {
		return nil, status.Error(codes.Unavailable, "ip geo not ready")
	}
	list, err := l.svcCtx.IpGeo.SearchList(l.ctx, in.GetIps(), in.GetLang(),
		geo.AccountOperatorCode(l.ctx))
	if err != nil {
		if errors.Is(err, geo.ErrTooManyIPs) || errors.Is(err, geo.ErrBadLang) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}
	return &integration.SearchIpResp{List: toPBList(list)}, nil
}
