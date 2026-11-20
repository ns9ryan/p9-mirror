package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/paging"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/activityi18n"
	"oa.98ent.com/p9/operator-activity/ent/activityorder"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// ErrOrderNotPending 订单不是待审核，不能通过或拒绝
var ErrOrderNotPending = errors.New("order not pending audit")

const (
	OrderTimePayout  int32 = 1 // 领取时间 payout_at
	OrderTimeApplied int32 = 2 // 申请时间 applied_at
)

// ListActivityOrderParams 活动订单列表查询参数
type ListActivityOrderParams struct {
	Page        int64
	PageSize    int64
	Lang        string
	OrderNo     string
	ActivityID  *int64
	MemberID    *int64
	Type        *int32
	Status      *int32
	AuditStatus *int32
	// TimeField 1 领取时间 payout_at，2 申请时间 applied_at；0 不加时间条件
	TimeField int32
	StartTime *time.Time
	EndTime   *time.Time
	// PendingOnly 为 true 时只查待审核，忽略 Status
	PendingOnly bool
}

// ActivityOrderRow 活动订单列表行
type ActivityOrderRow struct {
	ID            int64
	OrderNo       string
	ActivityID    int64
	ActivityName  string
	Type          int
	MemberID      int64
	RewardsTotal  int64
	ChannelSource int
	PayoutMode    int
	Status        int
	ClaimExpireAt *time.Time
	PayoutAt      *time.Time
	AppliedAt     *time.Time
	ApprovedAt    *time.Time
	ApproverUID   int64
}

// ApproveActivityOrderInput 审核通过后要落库的订单字段
type ApproveActivityOrderInput struct {
	ID           int64
	OrderNo      string
	Status       types.OrderStatus
	Remark       string
	ApproverUID  int64
	ApprovedAt   time.Time
	PayoutMode   int
	PayoutTiming int
	PayoutAt     time.Time
	ClaimExp     time.Time
}

// ListActivityOrders 分页查询活动订单，并补当前语言的活动名称
func (s *AdminService) ListActivityOrders(ctx context.Context, params ListActivityOrderParams) ([]ActivityOrderRow, int64, error) {
	offset, limit, ok := paging.OffsetLimit(params.Page, params.PageSize)
	if !ok {
		return nil, 0, errors.New("invalid page")
	}
	lang := strings.TrimSpace(params.Lang)

	q := s.db.ActivityOrder.Query()
	if orderNo := strings.TrimSpace(params.OrderNo); orderNo != "" {
		q = q.Where(activityorder.OrderNoEQ(orderNo))
	}
	if params.ActivityID != nil && *params.ActivityID > 0 {
		q = q.Where(activityorder.ActivityIDEQ(*params.ActivityID))
	}
	if params.MemberID != nil && *params.MemberID > 0 {
		q = q.Where(activityorder.MemberIDEQ(*params.MemberID))
	}
	if params.Type != nil && *params.Type > 0 {
		q = q.Where(activityorder.TypeEQ(int(*params.Type)))
	}
	if params.PendingOnly {
		q = q.Where(activityorder.StatusEQ(int(types.OrderStatusPendingAudit)))
	} else if params.Status != nil && *params.Status > 0 {
		q = q.Where(activityorder.StatusEQ(int(*params.Status)))
	}
	if params.AuditStatus != nil && *params.AuditStatus > 0 {
		q = q.Where(activityorder.AuditStatusEQ(int(*params.AuditStatus)))
	}
	q = applyOrderTimeRange(q, params.TimeField, params.StartTime, params.EndTime)

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []ActivityOrderRow{}, 0, nil
	}

	rows, err := q.
		Order(activityorder.ByID(sql.OrderDesc())).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	names, err := s.activityNames(ctx, rows, lang)
	if err != nil {
		return nil, 0, err
	}

	list := make([]ActivityOrderRow, 0, len(rows))
	for _, row := range rows {
		list = append(list, ActivityOrderRow{
			ID:            row.ID,
			OrderNo:       row.OrderNo,
			ActivityID:    row.ActivityID,
			ActivityName:  names[row.ActivityID],
			Type:          row.Type,
			MemberID:      row.MemberID,
			RewardsTotal:  row.RewardsTotal,
			ChannelSource: row.ChannelSource,
			PayoutMode:    row.PayoutMode,
			Status:        row.Status,
			ClaimExpireAt: row.ClaimExpireAt,
			PayoutAt:      row.PayoutAt,
			AppliedAt:     row.AppliedAt,
			ApprovedAt:    row.ApprovedAt,
			ApproverUID:   row.ApproverUID,
		})
	}
	return list, int64(total), nil
}

// ApproveActivityOrder 审核通过落库。订单不是待审核时返回 ErrOrderNotPending。
func (s *AdminService) ApproveActivityOrder(ctx context.Context, in ApproveActivityOrderInput) error {
	upd := s.db.ActivityOrder.Update().
		Where(
			activityorder.IDEQ(in.ID),
			activityorder.StatusEQ(int(types.OrderStatusPendingAudit)),
		).
		SetStatus(int(in.Status)).
		SetRemark(in.Remark).
		SetApprovedAt(in.ApprovedAt).
		SetApproverUID(in.ApproverUID).
		SetPayoutMode(in.PayoutMode).
		SetPayoutTiming(in.PayoutTiming).
		SetNillablePayoutAt(timePtr(in.PayoutAt)).
		SetNillableClaimExpireAt(timePtr(in.ClaimExp))
	return s.savePendingOrder(ctx, in.ID, upd)
}

// RejectActivityOrder 审核拒绝落库。订单不是待审核时返回 ErrOrderNotPending。
func (s *AdminService) RejectActivityOrder(ctx context.Context, id int64, remark string, approverUID int64, approvedAt time.Time) error {
	upd := s.db.ActivityOrder.Update().
		Where(
			activityorder.IDEQ(id),
			activityorder.StatusEQ(int(types.OrderStatusPendingAudit)),
		).
		SetStatus(int(types.OrderStatusRejected)).
		SetRemark(remark).
		SetApprovedAt(approvedAt).
		SetApproverUID(approverUID)
	return s.savePendingOrder(ctx, id, upd)
}

func (s *AdminService) savePendingOrder(ctx context.Context, id int64, upd *ent.ActivityOrderUpdate) error {
	n, err := upd.Save(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = s.db.ActivityOrder.Get(ctx, id)
	if err != nil {
		return err
	}
	return ErrOrderNotPending
}

func (s *AdminService) activityNames(ctx context.Context, rows []*ent.ActivityOrder, lang string) (map[int64]string, error) {
	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.ActivityID]; ok {
			continue
		}
		seen[row.ActivityID] = struct{}{}
		ids = append(ids, row.ActivityID)
	}
	names := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}
	i18nRows, err := s.db.ActivityI18n.Query().
		Where(activityi18n.ActivityIDIn(ids...), activityi18n.LangEQ(lang)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range i18nRows {
		if item == nil {
			continue
		}
		names[item.ActivityID] = item.Name
	}
	return names, nil
}

// applyOrderTimeRange 应用时间范围条件
func applyOrderTimeRange(q *ent.ActivityOrderQuery, timeField int32, start, end *time.Time) *ent.ActivityOrderQuery {
	switch timeField {
	case OrderTimePayout:
		if start != nil {
			q = q.Where(activityorder.PayoutAtGTE(*start))
		}
		if end != nil {
			q = q.Where(activityorder.PayoutAtLTE(*end))
		}
	case OrderTimeApplied:
		if start != nil {
			q = q.Where(activityorder.AppliedAtGTE(*start))
		}
		if end != nil {
			q = q.Where(activityorder.AppliedAtLTE(*end))
		}
	}
	return q
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
