package manage

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateLogic 编辑活动（仅草稿）
func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogic) Update(req *types.UpdateActivityReq) (resp *types.UpdateActivityResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.Update(l.ctx, toUpdateActivityReq(req))
	if err != nil {
		return nil, err
	}
	return &types.UpdateActivityResp{Id: rpcResp.GetId()}, nil
}
