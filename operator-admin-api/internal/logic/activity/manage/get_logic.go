package manage

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLogic 活动详情
func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLogic) Get(req *types.GetActivityReq) (resp *types.GetActivityResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.Get(l.ctx, &activitypb.GetActivityReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	detail := toActivityDetail(rpcResp.GetActivity())
	if detail == nil {
		return &types.GetActivityResp{}, nil
	}
	return &types.GetActivityResp{ActivityDetail: *detail}, nil
}
