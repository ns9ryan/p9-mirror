package activityservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 审核通过
	auditActionPass int32 = 1
	// 审核拒绝
	auditActionReject int32 = 2
)

type AuditOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuditOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditOrderLogic {
	return &AuditOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AuditOrder 活动订单审核：通过或拒绝
func (l *AuditOrderLogic) AuditOrder(in *activitypb.AuditActivityOrderReq) (*activitypb.AuditActivityOrderResp, error) {
	// 验证操作员身份
	claims, ok := ctxdata.OperatorClaimsFromCtx(l.ctx)
	if !ok {
		return nil, xerr.RpcErr(xerr.Unauthorized(i18n.Unauthorized))
	}
	orderNo := strings.TrimSpace(in.GetOrderNo())
	if orderNo == "" {
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}

	order, err := repository.New(l.svcCtx.DB).GetOrderByOrderNo(l.ctx, orderNo)
	if err != nil {
		return nil, auditOrderErr(err)
	}

	remark := strings.TrimSpace(in.GetRemark())
	adminServ := admin.NewService(l.svcCtx.DB)
	now := time.Now()
	switch in.GetAction() {
	case auditActionPass:
		if err := l.approve(adminServ, order, remark, claims.UserID, now); err != nil {
			return nil, err
		}
	case auditActionReject:
		if remark == "" {
			remark = "人工审核拒绝"
		}
		if err := adminServ.RejectActivityOrder(l.ctx, order.ID, remark, claims.UserID, now); err != nil {
			return nil, auditOrderErr(err)
		}
	default:
		return nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	return &activitypb.AuditActivityOrderResp{OrderNo: orderNo}, nil
}

// approve 审核通过逻辑处理
func (l *AuditOrderLogic) approve(adminServ *admin.AdminService, order *types.ActivityOrder, remark string, approverUID int64, now time.Time) error {
	if l.svcCtx.ActivityRuntime == nil {
		return xerr.RpcErr(xerr.BadRequest(i18n.ActivityEngineDisabled))
	}
	if order.Status != types.OrderStatusPendingAudit {
		return xerr.RpcErr(xerr.BadRequest(i18n.OrderNotPending))
	}
	activity, err := repository.New(l.svcCtx.DB).GetActivity(l.ctx, order.ActivityID)
	if err != nil {
		return auditOrderErr(err)
	}
	// 流水线工厂
	factory := l.svcCtx.ActivityRuntime.PipelineFactory(activity.Type)
	if factory == nil {
		return xerr.RpcErr(xerr.BadRequest(i18n.InvalidType))
	}
	// 构建发放流水线 奖励发放策略
	payout := factory.BuildPipeline(activity).Payout()
	if payout == nil {
		return xerr.RpcErr(xerr.BadRequest(i18n.InvalidType))
	}

	if remark == "" {
		remark = "人工审核通过"
	}
	order.Status = types.OrderStatusApproved
	order.Remark = remark
	order.ApprovedAt = now
	order.ApproverUID = approverUID
	payout.Apply(l.ctx, order)

	if err := adminServ.ApproveActivityOrder(l.ctx, admin.ApproveActivityOrderInput{
		ID:           order.ID,
		OrderNo:      order.OrderNo,
		Status:       order.Status,
		Remark:       order.Remark,
		ApproverUID:  order.ApproverUID,
		ApprovedAt:   order.ApprovedAt,
		PayoutMode:   int(order.PayoutMode),
		PayoutTiming: int(order.PayoutTiming),
		// PayoutAt:     order.PayoutAt, // 这时的PayoutAt是可领取时间代码内部判断使用，不能作为实际发放时间
		ClaimExp: order.ClaimExp,
	}); err != nil {
		return auditOrderErr(err)
	}
	// 执行奖励发放计划策略
	if err := payout.Schedule(l.ctx, order); err != nil {
		return xerr.RpcErr(err)
	}
	return nil
}

// auditOrderErr 审核订单错误处理
func auditOrderErr(err error) error {
	switch {
	case ent.IsNotFound(err):
		return xerr.RpcErr(xerr.EntNotFound(i18n.OrderNotFound, err))
	case errors.Is(err, admin.ErrOrderNotPending):
		return xerr.RpcErr(xerr.BadRequest(i18n.OrderNotPending))
	default:
		return xerr.RpcErr(err)
	}
}
