package condition

import (
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/metric"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// CycleModeCondition 周期模式条件
type CycleModeCondition struct {
	Manager *pipeline.Manager
}

func (c *CycleModeCondition) Name() string { return "cycle_mode" }

func (c *CycleModeCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	round, window, ok := c.Manager.CurrentRound(ctx.Activity, ctx.Event.GetEventTime())
	if !ok {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "cycle=%d event_time=%s out_of_cycle",
			ctx.Activity.CycleMode, ctx.Event.GetEventTime().Format("2006-01-02 15:04:05"))
		return pipeline.ResultSkip
	}
	ctx.Round = round
	ctx.Window = window
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "cycle=%d round=%d window=[%s,%s)",
		ctx.Activity.CycleMode, round,
		window.Start.Format("2006-01-02 15:04:05"),
		window.End.Format("2006-01-02 15:04:05"))
	return pipeline.ResultPass
}

// DepositTypeCondition 充值类型条件。指标由本条件自行计算。
type DepositTypeCondition struct {
	RewardType types.DepositRewardType
	Calc       *metric.DepositCalculator
}

func (c *DepositTypeCondition) Name() string { return "deposit_type" }

func (c *DepositTypeCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	m, ok := loadMetric(ctx, c.Name(), c.Calc)
	if !ok {
		return pipeline.ResultError
	}
	isFirstDeposit := m.DepositCount == 1  // 首充
	isSecondDeposit := m.DepositCount == 2 // 次存
	detail := "want_type=%d first=%v deposit_count=%d cumulative=%d amount=%d"
	switch c.RewardType {
	case types.DepositFirst: // 首充
		if isFirstDeposit {
			ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, detail+" matched=first",
				c.RewardType, isFirstDeposit, m.DepositCount, m.CumulativeDeposit, m.DepositAmount)
			return pipeline.ResultPass
		}
	case types.DepositSecond: // 次存
		if isSecondDeposit {
			ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, detail+" matched=second",
				c.RewardType, isSecondDeposit, m.DepositCount, m.CumulativeDeposit, m.DepositAmount)
			return pipeline.ResultPass
		}
	case types.DepositCumulative: // 累计
		if m.CumulativeDeposit > 0 {
			ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, detail+" matched=cumulative",
				c.RewardType, m.CumulativeDeposit, m.DepositCount, m.CumulativeDeposit, m.DepositAmount)
			return pipeline.ResultPass
		}
	case types.DepositSingle: // 单笔
		if m.DepositAmount > 0 {
			ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, detail+" matched=single",
				c.RewardType, m.DepositAmount, m.DepositCount, m.CumulativeDeposit, m.DepositAmount)
			return pipeline.ResultPass
		}
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, detail+" matched=none",
		c.RewardType, isFirstDeposit, m.DepositCount, m.CumulativeDeposit, m.DepositAmount)
	return pipeline.ResultSkip
}

// BetCondition 投注条件。指标由本条件自行计算。
type BetCondition struct {
	MinAmount      int64
	AllowedGameIds []int64
	Calc           *metric.BettingCalculator
}

func (c *BetCondition) Name() string { return "bet" }

func (c *BetCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	m, ok := loadMetric(ctx, c.Name(), c.Calc)
	if !ok {
		return pipeline.ResultError
	}
	bet := m.BetAmount
	gameID := ctx.Event.GetGameID()
	if bet < c.MinAmount {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "bet=%d min=%d game_id=%d allowed=%v",
			bet, c.MinAmount, gameID, c.AllowedGameIds)
		return pipeline.ResultSkip
	}
	// 这边指定游戏是指赠送奖励指定游戏洗码
	// if len(c.AllowedGameIds) > 0 {
	// 	if gameID == 0 || !slices.Contains(c.AllowedGameIds, gameID) {
	// 		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "bet=%d min=%d game_id=%d allowed=%v",
	// 			bet, c.MinAmount, gameID, c.AllowedGameIds)
	// 		return pipeline.ResultSkip
	// 	}
	// }
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "bet=%d min=%d game_id=%d allowed=%v",
		bet, c.MinAmount, gameID, c.AllowedGameIds)
	return pipeline.ResultPass
}

// ReferralCondition 推荐条件。指标由本条件自行计算。
type ReferralCondition struct {
	Schemas []types.RewardSchema
	Calc    *metric.ReferralCalculator
}

func (c *ReferralCondition) Name() string { return "referral" }

func (c *ReferralCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	if len(c.Schemas) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "schemas=0")
		return pipeline.ResultSkip
	}
	if _, ok := loadMetric(ctx, c.Name(), c.Calc); !ok {
		return pipeline.ResultError
	}
	minCount := c.Schemas[0].Count
	for _, s := range c.Schemas {
		if s.Count > 0 && s.Count < minCount {
			minCount = s.Count
		}
	}
	valid := pipeline.MetricValue(ctx)
	if minCount > 0 && valid < int64(minCount) {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "valid_referral=%d, min_count=%d", valid, minCount)
		return pipeline.ResultSkip
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "valid_referral=%d min_count=%d", valid, minCount)
	return pipeline.ResultPass
}

// WheelMetricCondition 转盘送机会前加载累计充值或累计投注。达标判断仍由调用方按档位完成。
type WheelMetricCondition struct {
	Calc *metric.WheelCalculator
}

func (c *WheelMetricCondition) Name() string { return "wheel_metric" }

func (c *WheelMetricCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	m, ok := loadMetric(ctx, c.Name(), c.Calc)
	if !ok {
		return pipeline.ResultError
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass,
		"deposit=%d cumulative_deposit=%d bet=%d cumulative_bet=%d",
		m.DepositAmount, m.CumulativeDeposit, m.BetAmount, m.CumulativeBet)
	return pipeline.ResultPass
}

func loadMetric(ctx *pipeline.EvalContext, name string, calc metric.Calculator) (*types.TriggerMetric, bool) {
	if calc == nil {
		ctx.StepLog.Cond(name, pipeline.ResultError, "calculator missing")
		return nil, false
	}
	m, err := calc.CalcMetric(ctx.Context, ctx.Activity, ctx.Event)
	if err != nil || m == nil {
		ctx.StepLog.Cond(name, pipeline.ResultError, "calc err=%v metric_nil=%v", err, m == nil)
		return nil, false
	}
	ctx.Metric = m
	return m, true
}
