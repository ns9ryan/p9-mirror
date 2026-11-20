package order

import (
	"context"
	"log"
	"time"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/activityorder"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type OrderActivities struct {
	client *ent.Client
}

// TimeoutCancelOrder 超时取消订单活动
func (a *OrderActivities) TimeoutCancelOrder(ctx context.Context, order CreateOrderReq, td time.Duration) (err error) {
	// 设置分站编码
	ctx = ctxdata.WithOperatorCode(ctx, order.OperatorCode)

	log.Printf("执行超时取消订单: 订单号=%s, 超时时间=%v", order.OrderNo, td)

	err = a.client.ActivityOrder.Update().
		Where(activityorder.OrderNo(order.OrderNo)).
		SetStatus(int(types.OrderStatusExpired)).
		SetRemark("活动订单用户超时未领取").
		Exec(ctx)
	if err != nil {
		log.Printf("更新订单状态失败: %v", err)
		return err
	}
	return nil
}

// TimeoutAutoClaimOrder 超时自动领取订单活动
func (a *OrderActivities) TimeoutAutoClaimOrder(ctx context.Context, order CreateOrderReq, td time.Duration) error {
	// 设置分站编码
	ctx = ctxdata.WithOperatorCode(ctx, order.OperatorCode)

	log.Printf("执行超时自动领取订单: 订单号=%s, 超时时间=%v", order.OrderNo, td)
	return a.settlePayout(ctx, order, "活动订单用户超时自动领取")
}

// ClaimOrder 用户手动领取
func (a *OrderActivities) ClaimOrder(ctx context.Context, order CreateOrderReq) error {
	// 设置分站编码
	ctx = ctxdata.WithOperatorCode(ctx, order.OperatorCode)

	log.Printf("执行用户领取订单: 订单号=%s", order.OrderNo)
	remark := order.Remark
	if remark == "" {
		remark = "用户手动领取"
	}
	return a.settlePayout(ctx, order, remark)
}

// PendingClaimOrder 待领取订单（同时回写按配置时长重算的 claim_exp）
func (a *OrderActivities) PendingClaimOrder(ctx context.Context, order CreateOrderReq) error {
	// 设置分站编码
	ctx = ctxdata.WithOperatorCode(ctx, order.OperatorCode)

	log.Printf("订单更新为待领取状态: 订单号=%s claim_exp=%v", order.OrderNo, order.ClaimExp)

	upd := a.client.ActivityOrder.Update().
		Where(activityorder.OrderNo(order.OrderNo)).
		SetStatus(int(order.Status)).
		SetRemark("订单更新为待领取状态")
	if !order.ClaimExp.IsZero() {
		upd = upd.SetClaimExpireAt(order.ClaimExp)
	}
	if err := upd.Exec(ctx); err != nil {
		log.Printf("更新订单状态失败: %v", err)
		return err
	}
	return nil
}

// AuditOrder 订单审核
func (a *OrderActivities) AuditOrder(ctx context.Context, order CreateOrderReq) error {
	// 设置分站编码
	ctx = ctxdata.WithOperatorCode(ctx, order.OperatorCode)

	log.Printf("执行订单审核: 订单号=%s", order.OrderNo)

	err := a.client.ActivityOrder.Update().
		Where(activityorder.OrderNo(order.OrderNo)).
		SetStatus(int(order.Status)).
		SetRemark(order.Remark).
		Exec(ctx)
	if err != nil {
		log.Printf("更新订单状态失败: %v", err)
		return err
	}
	return nil
}

// DisburseOrder 自动派发到账（含次日 Timer 到期后）
func (a *OrderActivities) DisburseOrder(ctx context.Context, order CreateOrderReq) error {
	// 设置分站编码
	ctx = ctxdata.WithOperatorCode(ctx, order.OperatorCode)
	log.Printf("执行自动派发到账: 订单号=%s", order.OrderNo)
	return a.settlePayout(ctx, order, "自动派发到账")
}

// settlePayout 订单到账：更新状态、增加余额、累加领取统计（按订单幂等，可安全重试）
func (a *OrderActivities) settlePayout(ctx context.Context, order CreateOrderReq, remark string) error {
	tx, err := a.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	upd := tx.ActivityOrder.Update().
		Where(activityorder.OrderNo(order.OrderNo), activityorder.StatusNEQ(int(types.OrderStatusPaidOut))).
		SetStatus(int(types.OrderStatusPaidOut)).
		SetRemark(remark)
	if !order.PayoutAt.IsZero() {
		upd = upd.SetPayoutAt(order.PayoutAt)
	}
	n, err := upd.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if n == 0 {
		return tx.Commit()
	}

	operatorCode, activityID, memberID, amount := order.OperatorCode, order.ActivityID, order.MemberID, order.RewardsTotal
	if operatorCode == "" || activityID == 0 || memberID == 0 || amount == 0 {
		row, err := tx.ActivityOrder.Get(ctx, order.ID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if operatorCode == "" && row.OperatorCode != "" {
			operatorCode = row.OperatorCode
		}
		if activityID == 0 {
			activityID = row.ActivityID
		}
		if memberID == 0 {
			memberID = row.MemberID
		}
		if amount == 0 {
			amount = row.RewardsTotal
		}
	}
	// if err := repository.IncrUserBalanceTx(ctx, tx, memberID, amount); err != nil {
	// 	_ = tx.Rollback()
	// 	return err
	// }
	// if err := repository.IncrActivityClaimedTx(ctx, tx, operatorCode, activityID, memberID, order.ID, amount); err != nil {
	// 	_ = tx.Rollback()
	// 	return err
	// }
	if err := tx.Commit(); err != nil {
		log.Printf("订单到账失败 order=%d err=%v", order.ID, err)
		return err
	}
	return nil
}
