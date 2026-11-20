package activityservicelogic

import (
	"time"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/admin"
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-activity/pkg/i18n"
)

func toActivityOrderListItem(row admin.ActivityOrderRow) *activitypb.ActivityOrderListItem {
	return &activitypb.ActivityOrderListItem{
		Id:               row.ID,
		OrderNo:          row.OrderNo,
		ActivityId:       row.ActivityID,
		ActivityName:     row.ActivityName,
		Type:             int32(row.Type),
		MemberId:         row.MemberID,
		RewardsTotal:     row.RewardsTotal,
		ChannelSource:    int32(row.ChannelSource),
		PayoutMode:       int32(row.PayoutMode),
		Status:           int32(row.Status),
		ClaimExpireAtSec: unixSecPtr(row.ClaimExpireAt),
		PayoutAtSec:      unixSecPtr(row.PayoutAt),
		AppliedAtSec:     unixSecPtr(row.AppliedAt),
		ApprovedAtSec:    unixSecPtr(row.ApprovedAt),
		ApproverUid:      row.ApproverUID,
	}
}

func toActivityOrderAuditListItem(row admin.ActivityOrderRow) *activitypb.ActivityOrderAuditListItem {
	return &activitypb.ActivityOrderAuditListItem{
		Id:            row.ID,
		OrderNo:       row.OrderNo,
		ActivityId:    row.ActivityID,
		ActivityName:  row.ActivityName,
		Type:          int32(row.Type),
		MemberId:      row.MemberID,
		RewardsTotal:  row.RewardsTotal,
		ChannelSource: int32(row.ChannelSource),
		PayoutMode:    int32(row.PayoutMode),
		Status:        int32(row.Status),
		AppliedAtSec:  unixSecPtr(row.AppliedAt),
	}
}

func orderTimeRange(start, end *int64) (startAt, endAt *time.Time, err error) {
	if start != nil && end != nil && *start > *end {
		return nil, nil, xerr.RpcErr(xerr.BadRequest(i18n.InvalidParam))
	}
	if start != nil {
		t := time.Unix(*start, 0)
		startAt = &t
	}
	if end != nil {
		t := time.Unix(*end, 0)
		endAt = &t
	}
	return startAt, endAt, nil
}

func unixSecPtr(t *time.Time) *int64 {
	if t == nil || t.IsZero() {
		return nil
	}
	sec := t.Unix()
	return &sec
}
