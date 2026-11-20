package activityservicelogic

import (
	"context"
	"strings"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAuditOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAuditOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuditOrdersLogic {
	return &ListAuditOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAuditOrders 活动订单审核列表
func (l *ListAuditOrdersLogic) ListAuditOrders(in *activitypb.ListActivityOrderAuditsReq) (*activitypb.ListActivityOrderAuditsResp, error) {
	if _, ok := ctxdata.OperatorClaimsFromCtx(l.ctx); !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}

	params := admin.ListActivityOrderParams{
		Page:        in.GetPage(),
		PageSize:    in.GetPageSize(),
		Lang:        strings.TrimSpace(in.GetLang()),
		OrderNo:     strings.TrimSpace(in.GetOrderNo()),
		PendingOnly: true,
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
	if in.AuditStatus != nil {
		v := in.GetAuditStatus()
		if v != int32(types.AuditStatusManual) && v != int32(types.AuditStatusRisk) {
			return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
		}
		params.AuditStatus = &v
	}
	if in.StartSec != nil || in.EndSec != nil {
		if in.TimeField == nil || in.GetTimeField() != admin.OrderTimeApplied {
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
	list := make([]*activitypb.ActivityOrderAuditListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toActivityOrderAuditListItem(row))
	}
	return &activitypb.ListActivityOrderAuditsResp{
		Total: total,
		List:  list,
	}, nil
}
