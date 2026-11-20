package manage

import (
	"context"

	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPublishLogic 发布活动
func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublishLogic) Publish(req *types.PublishActivityReq) (resp *types.PublishActivityResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.Publish(l.ctx, &activitypb.PublishActivityReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.PublishActivityResp{Id: rpcResp.GetId()}, nil
}
