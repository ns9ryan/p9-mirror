package repository

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"oa.98ent.com/p9/operator-activity/ent/activityorder"
)

// LoadInviterID 查询用户的邀请人 ID
func (r *Repository) LoadInviterID(userID int64) (int64, error) {
	// TODO: 实现查询用户的邀请人 ID
	return 0, nil
}

// CountValidReferrals 统计有效推荐人数（支持充值达标过滤） 返回: 邀请人数、到达人数、新有效推荐人数
func (r *Repository) CountValidReferrals(ctx context.Context, evt types.ActivityEvent, activity *types.Activity, window types.TimeWindow) (int, int, int, error) {
	// TODO: 实现统计有效推荐人数
	return 0, 0, 0, nil
}

// CountOrdersByPromo 统计活动已发奖励订单数
func (r *Repository) CountOrdersByPromo(ctx context.Context, activityID int64) (int, error) {
	n, err := r.client.ActivityOrder.Query().Where(activityorder.ActivityID(activityID)).Count(ctx)
	return n, err
}
