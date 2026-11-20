package repository

import (
	"context"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/activitystats"
	"oa.98ent.com/p9/operator-activity/pkg/database"
)

// ActivityStats 活动领取统计视图
type ActivityStats struct {
	OperatorCode      string
	ActivityID        int64
	ClaimMemberCount  int64
	ClaimedAmount     int64
	ClaimedOrderCount int64
}

// GetPromoStat 查询活动统计；无记录时返回零值
func (r *Repository) GetActivityStats(activityID int64) (*ActivityStats, error) {
	return GetActivityStats(r.client, activityID)
}

// GetPromoStat 按活动 ID 查询统计
func GetActivityStats(client *ent.Client, activityID int64) (*ActivityStats, error) {
	if client == nil {
		return &ActivityStats{ActivityID: activityID}, nil
	}
	row, err := client.ActivityStats.Query().Where(activitystats.ActivityID(activityID)).Only(context.Background())
	if ent.IsNotFound(err) {
		return &ActivityStats{ActivityID: activityID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &ActivityStats{
		OperatorCode:      row.OperatorCode,
		ActivityID:        row.ActivityID,
		ClaimMemberCount:  row.ClaimMemberCount,
		ClaimedAmount:     row.ClaimedAmount,
		ClaimedOrderCount: row.ClaimedOrderCount,
	}, nil
}

// IncrPromoClaimed 订单到账后累加领取人数/金额（按 orderID 幂等，可安全重试）
func (r *Repository) IncrPromoClaimed(ctx context.Context, partnerCode string, promoID, userID, orderID, amount int64) error {
	return IncrPromoClaimed(ctx, r.client, partnerCode, promoID, userID, orderID, amount)
}

// IncrPromoClaimed 供 Worker Activity 等直接使用 *ent.Client 调用
func IncrPromoClaimed(ctx context.Context, client *ent.Client, partnerCode string, promoID, userID, orderID, amount int64) error {
	if client == nil || promoID <= 0 || orderID <= 0 || amount <= 0 {
		return nil
	}
	return database.WithTx(ctx, client, func(tx *ent.Tx) error {
		return IncrPromoClaimedTx(ctx, tx, partnerCode, promoID, userID, orderID, amount)
	})
}

// IncrPromoClaimedTx 在已有事务中累加领取统计
func IncrPromoClaimedTx(ctx context.Context, tx *ent.Tx, partnerCode string, activityID, memberID, orderID, amount int64) error {
	if tx == nil || activityID <= 0 || memberID <= 0 || orderID <= 0 || amount <= 0 {
		return nil
	}

	if _, err := tx.ActivityStatsOrder.Create().
		SetActivityOrderID(orderID).
		SetOperatorCode(partnerCode).
		SetActivityID(activityID).
		SetMemberID(memberID).
		SetAmount(amount).
		Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return err
	}

	newUser := true
	if _, err := tx.ActivityStatsClaimer.Create().
		SetOperatorCode(partnerCode).
		SetActivityID(activityID).
		SetMemberID(memberID).
		Save(ctx); err != nil {
		if !ent.IsConstraintError(err) {
			return err
		}
		newUser = false
	}

	st, err := tx.ActivityStats.Query().
		Where(activitystats.ActivityID(activityID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		memberCnt := int64(0)
		if newUser {
			memberCnt = 1
		}
		_, err = tx.ActivityStats.Create().
			SetOperatorCode(partnerCode).
			SetActivityID(activityID).
			SetClaimMemberCount(memberCnt).
			SetClaimedAmount(amount).
			SetClaimedOrderCount(1).
			Save(ctx)
		if ent.IsConstraintError(err) {
			upd := tx.ActivityStats.Update().
				Where(activitystats.ActivityID(activityID)).
				AddClaimedAmount(amount).
				AddClaimedOrderCount(1)
			if newUser {
				upd = upd.AddClaimMemberCount(1)
			}
			return upd.Exec(ctx)
		}
		return err
	}
	if err != nil {
		return err
	}

	upd := tx.ActivityStats.UpdateOneID(st.ID).
		AddClaimedAmount(amount).
		AddClaimedOrderCount(1)
	if newUser {
		upd = upd.AddClaimMemberCount(1)
	}
	return upd.Exec(ctx)
}
