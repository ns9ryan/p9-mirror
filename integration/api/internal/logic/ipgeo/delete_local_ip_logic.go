package ipgeo

import (
	"context"

	"oa.98ent.com/p9/integration/api/internal/svc"
	"oa.98ent.com/p9/integration/api/internal/types"
	"oa.98ent.com/p9/integration/rpc/types/integration"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLocalIPLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLocalIPLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLocalIPLogic {
	return &DeleteLocalIPLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLocalIPLogic) DeleteLocalIP(req *types.IDReq) (*types.BaseMsgResp, error) {
	if _, err := l.svcCtx.Integration.DeleteLocalIp(l.ctx, &integration.IDReq{Id: req.Id}); err != nil {
		return nil, err
	}
	return &types.BaseMsgResp{Result: "success"}, nil
}
