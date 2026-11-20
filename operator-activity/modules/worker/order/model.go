// internal/model/order.go
package order

import (
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type CreateOrderReq struct {
	ID           int64              `json:"id"`
	OrderNo      string             `json:"order_no"`
	OperatorCode string             `json:"operator_code"`
	ActivityID   int64              `json:"activity_id"`
	MemberID     int64              `json:"member_id"`
	RewardsTotal int64              `json:"rewards_total"`
	ClaimExp     time.Time          `json:"claim_exp"`
	ApplyAt      time.Time          `json:"apply_at"`
	PayoutAt     time.Time          `json:"payout_at"`
	PayoutMode   types.PayoutMode   `json:"payout_mode"`
	PayoutTiming types.PayoutTiming `json:"payout_timing"`
	Status       types.OrderStatus  `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	Remark       string             `json:"remark"`
	CancelledAt  *time.Time         `json:"cancelled_at,omitempty"`
}
