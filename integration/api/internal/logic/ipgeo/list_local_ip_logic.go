package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/convert"
	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLocalIPLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLocalIPLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLocalIPLogic {
	return &ListLocalIPLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLocalIPLogic) ListLocalIP(req *types.LocalIPListReq) (*types.LocalIPListData, error) {
	out, err := l.svcCtx.Integration.ListLocalIps(l.ctx, &integration.LocalIPListReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Ip:       req.Ip,
	})
	if err != nil {
		return nil, err
	}
	return convert.LocalIPListData(out), nil
}
