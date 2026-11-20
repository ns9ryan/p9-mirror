package activityservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOrders 活动订单列表
func (l *ListOrdersLogic) ListOrders(in *activitypb.ListActivityOrdersReq) (*activitypb.ListActivityOrdersResp, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	params := admin.ListActivityOrderParams{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Lang:     strings.TrimSpace(in.GetLang()),
		OrderNo:  strings.TrimSpace(in.GetOrderNo()),
	}
	if in.ActivityId != nil {
		v := in.GetActivityId()
		params.ActivityID = &v
	}
	if in.MemberId != nil {
		v := in.GetMemberId()
		params.MemberID = &v
	}
	if in.Type != nil {
		v := in.GetType()
		params.Type = &v
	}
	if in.Status != nil {
		v := in.GetStatus()
		params.Status = &v
	}
	if in.StartSec != nil || in.EndSec != nil {
		if in.TimeField == nil || (in.GetTimeField() != admin.OrderTimePayout && in.GetTimeField() != admin.OrderTimeApplied) {
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
		}
		startAt, endAt, err := orderTimeRange(in.StartSec, in.EndSec)
		if err != nil {
			return nil, err
		}
		params.TimeField = in.GetTimeField()
		params.StartTime = startAt
		params.EndTime = endAt
	}

	rows, total, err := admin.NewService(l.svcCtx.DB).ListActivityOrders(l.ctx, params)
	if err != nil {
		return nil, xerr.RpcErr(err)
	}
	list := make([]*activitypb.ActivityOrderListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toActivityOrderListItem(row))
	}
	return &activitypb.ListActivityOrdersResp{
		Total: total,
		List:  list,
	}, nil
}
