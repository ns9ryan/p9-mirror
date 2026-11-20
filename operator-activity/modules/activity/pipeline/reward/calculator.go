package reward

import (
	"fmt"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// Calculator 奖励计算策略
type Calculator interface {
	Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error)
}

// BaseCalculator 公共逻辑
type BaseCalculator struct {
	Schemas []types.RewardSchema
}

func (b *BaseCalculator) idempotencyKey(ctx *pipeline.EvalContext, schemaID int64) string {
	settings := ctx.Activity.Settings
	// 充值活动需要特殊处理
	switch ctx.Activity.Type {
	case types.ActivityTypeDeposit:
		// 单次充值模式
		if settings.DepositRewardType == types.DepositSingle {
			switch ctx.Activity.CycleMode {
			case types.CycleInfinite: // 无限循环模式，幂等键需要包含订单ID
				return fmt.Sprintf("%d:%d:%d:%d:%d", ctx.Activity.ID, ctx.Member.ID, ctx.Round, schemaID, ctx.Event.GetOrderID())
			default:
				return fmt.Sprintf("%d:%d:%d:%d", ctx.Activity.ID, ctx.Member.ID, ctx.Round, schemaID)
			}
		}
	case types.ActivityTypeReferralMember:
		if settings.BonusMode == types.BonusSeparate {
			return fmt.Sprintf("%d:%d:%d:%d:%d_%d", ctx.Activity.ID, ctx.Member.ID, ctx.Round, schemaID, ctx.Event.GetEventType(), ctx.Event.GetOrderID())
		}
	case types.ActivityTypeWagering:
		return fmt.Sprintf("%d:%d:%d:%d:%d", ctx.Activity.ID, ctx.Member.ID, ctx.Round, schemaID, ctx.Event.GetOrderID())
	}
	// 其他模式下，幂等键只需要包含促销ID、用户ID、轮次和奖励ID
	return fmt.Sprintf("%d:%d:%d:%d", ctx.Activity.ID, ctx.Member.ID, ctx.Round, schemaID)
}
