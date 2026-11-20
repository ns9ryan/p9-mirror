package manage

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 创建活动
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateLogic) Create(req *types.CreateActivityReq) (resp *types.CreateActivityResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.Create(l.ctx, toCreateActivityReq(req))
	if err != nil {
		return nil, err
	}
	return &types.CreateActivityResp{Id: rpcResp.GetId()}, nil
}
