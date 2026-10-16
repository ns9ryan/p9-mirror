package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLocalIPLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLocalIPLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLocalIPLogic {
	return &CreateLocalIPLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateLocalIPLogic) CreateLocalIP(req *types.LocalIPView) (*types.LocalIPView, error) {
	out, err := l.svcCtx.Integration.CreateLocalIp(l.ctx, convert.LocalIP(req))
	if err != nil {
		return nil, err
	}
	view := convert.LocalIPView(out)
	return &view, nil
}
