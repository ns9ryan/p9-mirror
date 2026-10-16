package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLocalIPLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLocalIPLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLocalIPLogic {
	return &GetLocalIPLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLocalIPLogic) GetLocalIP(req *types.IDQuery) (*types.LocalIPView, error) {
	out, err := l.svcCtx.Integration.GetLocalIp(l.ctx, &integration.IDReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	view := convert.LocalIPView(out)
	return &view, nil
}
