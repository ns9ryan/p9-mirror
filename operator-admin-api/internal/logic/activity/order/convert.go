package order

import (
	"oa.98ent.com/p9/operator-activity/pb/operatoractivityrpc/activitypb"
	"oa.98ent.com/p9/operator-admin-api/internal/types"
)

func toActivityOrderListItem(item *activitypb.ActivityOrderListItem) types.ActivityOrderListItem {
	if item == nil {
		return types.ActivityOrderListItem{}
	}
	return types.ActivityOrderListItem{
		OrderNo:          item.GetOrderNo(),
		ActivityId:       item.GetActivityId(),
		ActivityName:     item.GetActivityName(),
		Type:             item.GetType(),
		MemberId:         item.GetMemberId(),
		RewardsTotal:     item.GetRewardsTotal(),
		ChannelSource:    item.GetChannelSource(),
		PayoutMode:       item.GetPayoutMode(),
		Status:           item.GetStatus(),
		ClaimExpireAtSec: item.ClaimExpireAtSec,
		PayoutAtSec:      item.PayoutAtSec,
		AppliedAtSec:     item.AppliedAtSec,
		ApprovedAtSec:    item.ApprovedAtSec,
		ApproverUid:      item.GetApproverUid(),
	}
}

func toActivityOrderAuditListItem(item *activitypb.ActivityOrderAuditListItem) types.ActivityOrderAuditListItem {
	if item == nil {
		return types.ActivityOrderAuditListItem{}
	}
	return types.ActivityOrderAuditListItem{
		OrderNo:       item.GetOrderNo(),
		ActivityId:    item.GetActivityId(),
		ActivityName:  item.GetActivityName(),
		Type:          item.GetType(),
		MemberId:      item.GetMemberId(),
		RewardsTotal:  item.GetRewardsTotal(),
		ChannelSource: item.GetChannelSource(),
		PayoutMode:    item.GetPayoutMode(),
		Status:        item.GetStatus(),
		AppliedAtSec:  item.AppliedAtSec,
	}
}
