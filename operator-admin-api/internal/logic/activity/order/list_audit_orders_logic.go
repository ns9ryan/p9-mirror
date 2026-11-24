package order

import (
	"context"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAuditOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListAuditOrdersLogic 活动订单审核列表
func NewListAuditOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuditOrdersLogic {
	return &ListAuditOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAuditOrdersLogic) ListAuditOrders(req *types.ListActivityOrderAuditsReq) (resp *types.ListActivityOrderAuditsResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.ListAuditOrders(l.ctx, &activitypb.ListActivityOrderAuditsReq{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Lang:        i18n.Lang(l.ctx, nil),
		OrderNo:     req.OrderNo,
		ActivityId:  req.ActivityId,
		MemberCode:  req.MemberCode,
		Type:        req.Type,
		AuditStatus: req.AuditStatus,
		TimeField:   req.TimeField,
		StartSec:    req.StartSec,
		EndSec:      req.EndSec,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityOrderAuditListItem, 0, len(rpcResp.GetList()))
	for _, item := range rpcResp.GetList() {
		list = append(list, toActivityOrderAuditListItem(item))
	}
	return &types.ListActivityOrderAuditsResp{
		Total: rpcResp.GetTotal(),
		List:  list,
	}, nil
}
