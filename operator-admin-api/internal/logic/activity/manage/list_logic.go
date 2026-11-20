package manage

import (
	"context"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic 活动列表
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLogic) List(req *types.ListActivitiesReq) (resp *types.ListActivitiesResp, err error) {
	rpcReq := &activitypb.ListActivitiesReq{
		Page:     req.Page,
		PageSize: req.PageSize,
		Name:     req.Name,
		Lang:     i18n.Lang(l.ctx, nil),
		Type:     req.Type,
		Client:   req.Client,
		Status:   req.Status,
	}
	rpcResp, err := l.svcCtx.ActivityRpc.List(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityListItem, 0, len(rpcResp.GetList()))
	for _, item := range rpcResp.GetList() {
		list = append(list, toActivityListItem(item))
	}
	return &types.ListActivitiesResp{
		Total: rpcResp.GetTotal(),
		List:  list,
	}, nil
}
