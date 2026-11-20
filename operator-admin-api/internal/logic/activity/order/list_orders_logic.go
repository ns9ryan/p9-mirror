package order

import (
	"context"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListOrdersLogic 活动订单列表
func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrdersLogic) ListOrders(req *types.ListActivityOrdersReq) (resp *types.ListActivityOrdersResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.ListOrders(l.ctx, &activitypb.ListActivityOrdersReq{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Lang:       i18n.Lang(l.ctx, nil),
		OrderNo:    req.OrderNo,
		ActivityId: req.ActivityId,
		MemberId:   req.MemberId,
		Type:       req.Type,
		Status:     req.Status,
		TimeField:  req.TimeField,
		StartSec:   req.StartSec,
		EndSec:     req.EndSec,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityOrderListItem, 0, len(rpcResp.GetList()))
	for _, item := range rpcResp.GetList() {
		list = append(list, toActivityOrderListItem(item))
	}
	return &types.ListActivityOrdersResp{
		Total: rpcResp.GetTotal(),
		List:  list,
	}, nil
}
