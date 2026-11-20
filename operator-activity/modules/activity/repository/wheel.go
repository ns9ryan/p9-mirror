package repository

import (
	"context"
	"fmt"
	"time"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/activitywheelchancebalance"
	"oa.98ent.com/p9/operator-activity/ent/activitywheelchancegrant"
	"oa.98ent.com/p9/operator-activity/pkg/database"
)

// ExistsWheelChanceGrant 机会发放幂等检查
func (r *Repository) ExistsWheelChanceGrant(ctx context.Context, key string) (bool, error) {
	return r.client.ActivityWheelChanceGrant.Query().
		Where(activitywheelchancegrant.IdempotencyKeyEQ(key)).
		Exist(ctx)
}

// GrantWheelChances 发放抽奖机会（事务：写流水 + 累加余额）
func (r *Repository) GrantWheelChances(ctx context.Context, m *WheelChanceGrantModel) error {
	if m == nil || m.Chances <= 0 {
		return nil
	}
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		c := tx.ActivityWheelChanceGrant.Create().
			SetOperatorCode(m.OperatorCode).
			SetActivityID(m.ActivityID).
			SetMemberID(m.MemberID).
			SetRound(m.Round).
			SetSchemaID(m.SchemaID).
			SetEventType(int(m.EventType)).
			SetOrderID(m.OrderID).
			SetChances(m.Chances).
			SetIdempotencyKey(m.IdempotencyKey)
		if !m.CreatedAt.IsZero() {
			c = c.SetCreatedAt(m.CreatedAt)
		}
		if _, err := c.Save(ctx); err != nil {
			if ent.IsConstraintError(err) {
				return nil
			}
			return err
		}

		bal, err := tx.ActivityWheelChanceBalance.Query().
			Where(
				activitywheelchancebalance.ActivityID(m.ActivityID),
				activitywheelchancebalance.MemberID(m.MemberID),
			).
			Only(ctx)
		if ent.IsNotFound(err) {
			_, err = tx.ActivityWheelChanceBalance.Create().
				SetOperatorCode(m.OperatorCode).
				SetActivityID(m.ActivityID).
				SetMemberID(m.MemberID).
				SetGranted(m.Chances).
				SetUsed(0).
				Save(ctx)
			if ent.IsConstraintError(err) {
				return tx.ActivityWheelChanceBalance.Update().
					Where(
						activitywheelchancebalance.ActivityID(m.ActivityID),
						activitywheelchancebalance.MemberID(m.MemberID),
					).
					AddGranted(m.Chances).
					Exec(ctx)
			}
			return err
		}
		if err != nil {
			return err
		}
		return tx.ActivityWheelChanceBalance.UpdateOneID(bal.ID).
			AddGranted(m.Chances).
			SetUpdatedAt(time.Now()).
			Exec(ctx)
	})
}

// GetWheelChanceBalance 查询剩余抽奖次数
func (r *Repository) GetWheelChanceBalance(ctx context.Context, activityID, memberID int64) (remaining int, err error) {
	bal, err := r.client.ActivityWheelChanceBalance.Query().
		Where(
			activitywheelchancebalance.ActivityID(activityID),
			activitywheelchancebalance.MemberID(memberID),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	remaining = bal.Granted - bal.Used
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// ConsumeWheelChance 消耗一次抽奖机会；不足返回错误
func (r *Repository) ConsumeWheelChance(ctx context.Context, activityID, memberID int64) error {
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		bal, err := tx.ActivityWheelChanceBalance.Query().
			Where(
				activitywheelchancebalance.ActivityID(activityID),
				activitywheelchancebalance.MemberID(memberID),
			).
			Only(ctx)
		if ent.IsNotFound(err) {
			return fmt.Errorf("no wheel chance")
		}
		if err != nil {
			return err
		}
		if bal.Granted-bal.Used < 1 {
			return fmt.Errorf("no wheel chance")
		}
		return tx.ActivityWheelChanceBalance.UpdateOneID(bal.ID).
			AddUsed(1).
			SetUpdatedAt(time.Now()).
			Exec(ctx)
	})
}
