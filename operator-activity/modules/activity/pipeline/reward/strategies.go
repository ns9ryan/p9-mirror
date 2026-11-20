package reward

import (
	"fmt"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// FixedCalculator 固定金额
type FixedCalculator struct {
	BaseCalculator
}

// Calculate 固定金额（合并多个达标档）
// func (c *FixedCalculator) Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error) {
// 	var (
// 		total   int64
// 		bestAmt int64
// 		bestID  int64
// 		ids     []string
// 	)
// 	for _, schema := range c.Schemas {
// 		if !matchThreshold(ctx, schema) {
// 			continue
// 		}
// 		total += schema.Amount
// 		ids = append(ids, fmt.Sprintf("%d", schema.ID))
// 		if bestID == 0 || schema.Amount > bestAmt {
// 			bestAmt = schema.Amount
// 			bestID = schema.ID
// 		}
// 	}
// 	if len(ids) == 0 {
// 		return nil, nil
// 	}
// 	return []types.RewardItem{{
// 		SchemaID:       bestID,
// 		Amount:         total,
// 		Currency:       ctx.Event.GetCurrency(),
// 		Remark:         fmt.Sprintf("fixed reward merged schemas=[%s]", strings.Join(ids, ",")),
// 		IdempotencyKey: c.idempotencyKey(ctx, bestID),
// 	}}, nil
// }

// Calculate 固定金额（不合并下发多个达标档）
func (c *FixedCalculator) Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error) {
	var items []types.RewardItem
	for _, schema := range c.Schemas {
		// 过滤不达标档或本次之前已达标档
		if !pipeline.MatchThreshold(ctx, schema) || pipeline.FilterThreshold(ctx, schema) {
			continue
		}
		items = append(items, types.RewardItem{
			SchemaID:       schema.ID,
			Amount:         schema.Amount,
			Remark:         fmt.Sprintf("fixed reward schema=%d", schema.ID),
			IdempotencyKey: c.idempotencyKey(ctx, schema.ID),
		})
	}
	return items, nil
}

// PercentCalculator 按比例
type PercentCalculator struct {
	BaseCalculator
}

func (c *PercentCalculator) Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error) {
	var items []types.RewardItem
	for _, schema := range c.Schemas {
		if !pipeline.MatchThreshold(ctx, schema) || pipeline.FilterThreshold(ctx, schema) {
			continue
		}
		amount := ctx.Event.GetAmount() * int64(schema.Rate)
		if schema.MaxBonus > 0 && amount > schema.MaxBonus {
			amount = schema.MaxBonus
		}
		if amount <= 0 {
			continue
		}
		items = append(items, types.RewardItem{
			SchemaID:       schema.ID,
			Amount:         amount,
			Remark:         fmt.Sprintf("percent reward rate=%d", schema.Rate),
			IdempotencyKey: c.idempotencyKey(ctx, schema.ID),
		})
	}
	return items, nil
}

// FixedTieredCalculator 固定金额阶梯最高下发
type FixedTieredCalculator struct {
	BaseCalculator
}

func (c *FixedTieredCalculator) Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error) {
	value := pipeline.MetricValue(ctx)
	var matched *types.RewardSchema
	for i := range c.Schemas {
		s := &c.Schemas[i]
		if value >= s.Threshold {
			if matched == nil || s.Threshold > matched.Threshold {
				matched = s
			}
		}
	}
	if matched == nil || pipeline.FilterThreshold(ctx, *matched) {
		return nil, nil
	}
	return []types.RewardItem{{
		SchemaID:       matched.ID,
		Amount:         matched.Amount,
		Remark:         fmt.Sprintf("fixed tiered reward threshold=%d amount=%d", matched.Threshold, matched.Amount),
		IdempotencyKey: c.idempotencyKey(ctx, matched.ID),
	}}, nil
}

// PercentageTieredCalculator 按比例阶梯最高下发
type PercentageTieredCalculator struct {
	BaseCalculator
}

func (c *PercentageTieredCalculator) Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error) {
	var matched *types.RewardSchema
	for i := range c.Schemas {
		s := &c.Schemas[i]
		if pipeline.MetricValue(ctx) >= s.Threshold {
			if matched == nil || s.Threshold > matched.Threshold {
				matched = s
			}
		}
	}
	if matched == nil || pipeline.FilterThreshold(ctx, *matched) {
		return nil, nil
	}
	return []types.RewardItem{{
		SchemaID:       matched.ID,
		Amount:         ctx.Event.GetAmount() * int64(matched.Rate),
		Remark:         fmt.Sprintf("percentage tiered reward threshold=%d rate=%d", matched.Threshold, matched.Rate),
		IdempotencyKey: c.idempotencyKey(ctx, matched.ID),
	}}, nil
}

// SeparateCalculator 单独奖励，如推荐活动，单独人数奖励
type SeparateCalculator struct {
	BaseCalculator
}

func (c *SeparateCalculator) Calculate(ctx *pipeline.EvalContext) ([]types.RewardItem, error) {
	if len(c.Schemas) == 0 {
		return nil, nil
	}
	schema := c.Schemas[0]

	if pipeline.MetricValue(ctx) < schema.Threshold {
		return nil, nil
	}

	return []types.RewardItem{{
		SchemaID:       schema.ID,
		Amount:         schema.Amount,
		Remark:         "separate reward",
		IdempotencyKey: c.idempotencyKey(ctx, schema.ID),
	}}, nil
}
