package manage

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCloseLogic 关闭活动
func NewCloseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseLogic {
	return &CloseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseLogic) Close(req *types.CloseActivityReq) (resp *types.CloseActivityResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.Close(l.ctx, &activitypb.CloseActivityReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.CloseActivityResp{Id: rpcResp.GetId()}, nil
}
