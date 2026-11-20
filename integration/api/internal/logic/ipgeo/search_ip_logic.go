package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchIpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchIpLogic {
	return &SearchIpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchIpLogic) SearchIp(req *types.SearchIpReq) (*types.SearchIpResp, error) {
	out, err := l.svcCtx.Integration.SearchIp(l.ctx, &integration.SearchIpReq{
		Ips:  req.Ips,
		Lang: req.Lang,
	})
	if err != nil {
		return nil, err
	}
	return convert.SearchIpResp(out), nil
}
