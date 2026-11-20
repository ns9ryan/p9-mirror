package repository

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/condition"
	"oa.98ent.com/p9/operator-activity/pkg/database"

	"oa.98ent.com/p9/operator-activity/ent"

	"oa.98ent.com/p9/operator-activity/ent/activity"
	"oa.98ent.com/p9/operator-activity/ent/activityapplication"
	"oa.98ent.com/p9/operator-activity/ent/activityorder"
	"oa.98ent.com/p9/operator-activity/ent/activityorderitem"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"oa.98ent.com/p9/common/ctxdata"
	activitycache "oa.98ent.com/p9/operator-activity/modules/activity/repository/cache"
)

// Repository 仓库
type Repository struct {
	client    *ent.Client
	publisher *KafkaPublisher
}

func New(client *ent.Client, brokers ...string) *Repository {
	return &Repository{
		client:    client,
		publisher: NewKafkaPublisher(brokers),
	}
}

// 加载进行中活动 优先缓存获取，缓存不存在则从数据库加载
func (r *Repository) LoadOngoingActivities(ctx context.Context, activityType types.ActivityType) ([]*types.Activity, error) {
	operatorCode := ctxdata.OperatorCodeFromCtx(ctx)
	// 上下文中获取不到分站编码
	if operatorCode == "" {
		return nil, fmt.Errorf("context missing operator code")
	}
	return activitycache.GetOngoingActivitiesCache(operatorCode, activityType, func(_ map[string]any) (any, error) {
		return r.loadOngoingPromotionsFromDB(ctx, activityType)
	})
}

// 删除指定类型活动的缓存
func (r *Repository) DelOngoingActivitiesCache(operatorCode string, activityType types.ActivityType) {
	activitycache.DelOngoingActivitiesCache(operatorCode, activityType)
}

// 从数据库加载进行中活动
func (r *Repository) loadOngoingPromotionsFromDB(ctx context.Context, activityType types.ActivityType) ([]*types.Activity, error) {
	rows, err := r.client.Activity.Query().
		Where(
			activity.TypeEQ(int(activityType)),
			activity.StatusEQ(int(types.ActivityStatusOngoing)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*types.Activity, 0, len(rows))
	for _, row := range rows {
		p, err := toActivity(ActivityFromEnt(row))
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

// 根据活动ID查找活动
func (r *Repository) GetActivity(ctx context.Context, activityID int64) (*types.Activity, error) {
	row, err := r.client.Activity.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	return toActivity(ActivityFromEnt(row))
}

// 判断用户是否已申请活动
func (r *Repository) HasApplication(ctx context.Context, activityID, memberID int64) (bool, error) {
	n, err := r.client.ActivityApplication.Query().
		Where(activityapplication.ActivityID(activityID), activityapplication.MemberID(memberID)).
		Count(ctx)
	return n > 0, err
}

// 默认参与活动
func (r *Repository) DefaultApplication(ctx context.Context, activityID, memberID int64) error {
	_, err := r.client.ActivityApplication.Create().
		SetActivityID(activityID).
		SetMemberID(memberID).
		SetApplyMode(int(types.ApplyAuto)).
		SetApplyAt(time.Now()).
		Save(ctx)
	return err
}

// 判断幂等键是否已存在
func (r *Repository) IdempKeyExists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.ActivityOrderItem.Query().
		Where(activityorderitem.IdempotencyKeyEQ(key)).
		Count(ctx)
	return n > 0, err
}

// 统计订单数量
func (r *Repository) CountOrders(ctx context.Context, req condition.LimitCountRequest) (int, error) {
	q := r.client.ActivityOrder.Query().Where(activityorder.ActivityID(req.ActivityID))
	switch req.Scope {
	case "per_user":
		q = q.Where(activityorder.MemberID(req.MemberID))
	case "per_cycle":
		q = q.Where(activityorder.MemberID(req.MemberID), activityorder.RoundEQ(req.Round))
	case "per_day":
		q = q.Where(activityorder.MemberID(req.MemberID), activityorder.CreatedAtGTE(req.Window.Start), activityorder.CreatedAtLT(req.Window.End))
	case "total":
	default:
		q = q.Where(activityorder.MemberID(req.MemberID))
	}
	n, err := q.Count(ctx)
	return n, err
}

// 统计申请活动数量
func (r *Repository) CountApplications(ctx context.Context, req condition.LimitCountRequest) (int, error) {
	q := r.client.ActivityApplication.Query().Where(activityapplication.ActivityID(req.ActivityID))
	if req.Scope != "total" {
		q = q.Where(activityapplication.MemberID(req.MemberID))
	}
	n, err := q.Count(ctx)
	return n, err
}

// 统计审核通过充值数量
func (r *Repository) CountApprovedDeposits(ctx context.Context, memberID int64, window types.TimeWindow) (int, error) {
	// TODO: 实现统计审核通过充值数量
	return 0, nil
}

// 统计审核通过充值金额
func (r *Repository) SumDeposits(ctx context.Context, memberID int64, window types.TimeWindow) (int64, error) {
	// TODO: 实现统计审核通过充值金额
	return 0, nil
}

// 统计审核通过投注金额
func (r *Repository) SumBets(ctx context.Context, memberID int64, window types.TimeWindow) (int64, error) {
	// TODO: 实现统计审核通过投注金额
	return 0, nil
}

// 统计审核通过投注数量
func (r *Repository) CountBets(ctx context.Context, memberID int64, window types.TimeWindow) (int64, error) {
	// TODO: 实现统计审核通过投注数量
	return 0, nil
}

// 保存活动奖励订单
func (r *Repository) SaveOrders(ctx context.Context, orders []*types.ActivityOrder, activity *types.Activity, user *types.MemberInfo) error {
	err := database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		for _, order := range orders {
			c := tx.ActivityOrder.Create().
				SetOrderNo(order.OrderNo).
				SetMemberID(order.MemberID).
				SetAgentID(order.AgentID).
				SetActivityID(activity.ID).
				SetType(int(activity.Type)).
				SetRound(order.Round).
				SetEventType(int(order.EventType)).
				SetRelEventOrderID(order.RelOrderID).
				SetRelAmount(order.RelAmount).
				SetStatus(int(order.Status)).
				SetAuditStatus(int(order.AuditStatus)).
				SetApproverUID(order.ApproverUID).
				SetRewardsTotal(order.RewardsTotal).
				SetPayoutMode(int(order.PayoutMode)).
				SetPayoutTiming(int(order.PayoutTiming)).
				SetChannelSource(int(order.ClientType)).
				SetSubOrderCount(order.SubOrderCount).
				SetRemark(order.Remark).
				SetVersion(order.Version).
				SetAppliedAt(order.AppliedAt).
				SetNillableApprovedAt(approvedAtPtr(order.ApprovedAt)).
				SetUpdatedAt(order.CreatedAt).
				SetCreatedAt(order.CreatedAt)
			insertRow, err := c.Save(ctx)
			if err != nil {
				return err
			}
			// 赋值订单ID
			order.ID = insertRow.ID
			for _, orderItem := range order.Items {
				ic := tx.ActivityOrderItem.Create().
					SetActivityOrderID(order.ID).
					SetMemberID(orderItem.MemberID).
					SetActivityID(orderItem.ActivityID).
					SetActivityType(int(orderItem.ActivityType)).
					SetActivityRound(orderItem.Round).
					SetRewardAmount(orderItem.RewardAmount).
					SetIdempotencyKey(orderItem.IdempotencyKey).
					SetUpdatedAt(orderItem.CreatedAt).
					SetCreatedAt(orderItem.CreatedAt)
				if _, err := ic.Save(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if ent.IsConstraintError(err) {
		return nil
	}
	return err
}

// 根据订单ID查找订单
func (r *Repository) GetOrder(ctx context.Context, orderID int64) (*types.ActivityOrder, error) {
	row, err := r.client.ActivityOrder.Get(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return toActivityOrder(activityOrderFromEnt(row)), nil
}

// GetOrderByOrderNo 根据订单号查找订单
func (r *Repository) GetOrderByOrderNo(ctx context.Context, orderNo string) (*types.ActivityOrder, error) {
	row, err := r.client.ActivityOrder.Query().
		Where(activityorder.OrderNo(orderNo)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toActivityOrder(activityOrderFromEnt(row)), nil
}

// 更新订单派发状态
func (r *Repository) UpdateOrderPayout(ctx context.Context, order *types.ActivityOrder) error {
	return r.client.ActivityOrder.UpdateOneID(order.ID).
		SetStatus(int(order.Status)).
		SetRemark(order.Remark).
		SetPayoutMode(int(order.PayoutMode)).
		SetPayoutTiming(int(order.PayoutTiming)).
		Exec(ctx)
}

// 更新订单状态
func (r *Repository) UpdateOrderStatus(ctx context.Context, orderID int64, status types.OrderStatus, remark string) error {
	return r.client.ActivityOrder.UpdateOneID(orderID).
		SetStatus(int(status)).
		SetRemark(remark).
		Exec(ctx)
}

// 保存派发计划
func (r *Repository) SaveSchedule(ctx context.Context, order *types.ActivityOrder) error {
	_, err := r.client.ActivityPayoutSchedule.Create().
		SetID(order.ID).
		SetActivityID(order.ActivityID).
		SetActivityOrderID(order.ID).
		SetMemberID(order.MemberID).
		SetRewardsAmount(order.RewardsTotal).
		SetPayoutAt(order.PayoutAt).
		SetStatus(1).
		Save(ctx)
	return err
}

// 向消息中间件发送已派发奖励消息
func (r *Repository) PublishPayout(ctx context.Context, order *types.ActivityOrder) error {
	return r.publisher.PublishPayout(ctx, order)
}

// 向消息中间件发送派发计划创建消息
func (r *Repository) PublishOrderCreated(ctx context.Context, orders []*types.ActivityOrder) error {
	return nil
}

// 向消息中间件发送充值审核通过消息
func (r *Repository) PublishDepositApproved(ctx context.Context, payload *types.DepositPayload) error {
	return r.publisher.PublishDepositApproved(ctx, payload)
}

// 向消息中间件发送会员注册消息
func (r *Repository) PublishRegister(ctx context.Context, payload *types.RegisterPayload) error {
	return r.publisher.PublishRegister(ctx, payload)
}

// 向消息中间件发送投注结算消息
func (r *Repository) PublishBetSettled(ctx context.Context, payload *types.BetPayload) error {
	return r.publisher.PublishBetSettled(ctx, payload)
}

func toActivityOrder(m *ActivityOrderModel) *types.ActivityOrder {
	o := &types.ActivityOrder{
		ID:           m.ID,
		OrderNo:      m.OrderNo,
		OperatorCode: m.OperatorCode,
		MemberID:     m.MemberID,
		AgentID:      m.AgentID,
		ActivityID:   m.ActivityID,
		ActivityType: types.ActivityType(m.ActivityType),
		Round:        m.Round,
		EventType:    types.EventType(m.EventType),
		RelOrderID:   m.RelOrderID,
		RelAmount:    m.RelAmount,
		Status:       types.OrderStatus(m.Status),
		AuditStatus:  types.AuditStatus(m.AuditStatus),
		ApproverUID:  m.ApproverUID,
		RewardsTotal: m.RewardsTotal,
		PayoutMode:   types.PayoutMode(m.PayoutMode),
		PayoutTiming: types.PayoutTiming(m.PayoutTiming),
		Remark:       m.Remark,
		AppliedAt:    m.AppliedAt,
		CreatedAt:    m.CreatedAt,
	}
	if m.PayoutAt != nil {
		o.PayoutAt = *m.PayoutAt
	}
	if m.ClaimExp != nil {
		o.ClaimExp = *m.ClaimExp
	}
	if m.ApprovedAt != nil {
		o.ApprovedAt = *m.ApprovedAt
	}
	return o
}

func approvedAtPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
