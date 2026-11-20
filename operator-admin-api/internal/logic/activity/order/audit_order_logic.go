package order

import (
	"context"

	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/svc"
	"oa.98ent.com/p9/operator-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAuditOrderLogic 活动订单审核（通过或拒绝）
func NewAuditOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditOrderLogic {
	return &AuditOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuditOrderLogic) AuditOrder(req *types.AuditActivityOrderReq) (resp *types.AuditActivityOrderResp, err error) {
	rpcResp, err := l.svcCtx.ActivityRpc.AuditOrder(l.ctx, &activitypb.AuditActivityOrderReq{
		OrderNo: req.OrderNo,
		Action:  req.Action,
		Remark:  req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.AuditActivityOrderResp{OrderNo: rpcResp.GetOrderNo()}, nil
}
