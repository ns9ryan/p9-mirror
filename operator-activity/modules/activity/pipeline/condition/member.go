package condition

import (
	"context"
	"slices"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// FixedLevelCondition 固定层级条件
type FixedLevelCondition struct {
	Allowed []int64
}

// Name 返回条件名称
func (c *FixedLevelCondition) Name() string { return "fixed_level" }

// Evaluate 评估条件
func (c *FixedLevelCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	if len(c.Allowed) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "no_limit")
		return pipeline.ResultPass
	}
	if ctx.Member == nil {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "user=nil")
		return pipeline.ResultSkip
	}
	// 如果允许的固定层级为空，则返回 pass
	if len(c.Allowed) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "fixed_levels_no_limit, allowed=%v", c.Allowed)
		return pipeline.ResultPass
	}
	// 遍历固定层级，如果有一个符合条件，则返回 pass
	fixedLevels := ctx.Member.FixedLevels
	for _, fixedLevel := range fixedLevels {
		if slices.Contains(c.Allowed, fixedLevel) {
			ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "fixed_levels=%v allowed=%v", fixedLevels, c.Allowed)
			return pipeline.ResultPass
		}
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "fixed_levels=%v allowed=%v", fixedLevels, c.Allowed)
	return pipeline.ResultSkip
}

// AutoLevelCondition 自动层级条件
type AutoLevelCondition struct {
	Allowed []int64
}

// Name 返回条件名称
func (c *AutoLevelCondition) Name() string { return "auto_level" }

// Evaluate 评估条件
func (c *AutoLevelCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	if len(c.Allowed) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "no_limit")
		return pipeline.ResultPass
	}
	if ctx.Member == nil {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "user=nil")
		return pipeline.ResultSkip
	}
	autoLevel := ctx.Member.AutoLevel
	if !slices.Contains(c.Allowed, autoLevel) {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "auto_level=%d allowed=%v", autoLevel, c.Allowed)
		return pipeline.ResultSkip
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "auto_level=%d allowed=%v", autoLevel, c.Allowed)
	return pipeline.ResultPass
}

// ClientTypeCondition 客户端类型条件
type ClientTypeCondition struct {
	Allowed []int64
}

// Name 返回条件名称
func (c *ClientTypeCondition) Name() string { return "client_type" }

// Evaluate 评估条件
func (c *ClientTypeCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	if len(c.Allowed) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "no_limit")
		return pipeline.ResultPass
	}
	clientType := ctx.Event.GetClientType()
	if !slices.Contains(c.Allowed, int64(clientType)) {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "client_type=%d allowed=%v", clientType, c.Allowed)
		return pipeline.ResultSkip
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "client_type=%d allowed=%v", clientType, c.Allowed)
	return pipeline.ResultPass
}

// ApplyModeCondition 参与模式条件
type ApplyModeCondition struct {
	Mode types.ApplyMode
	Repo ApplicationChecker
}

// ApplicationChecker 参与记录检查
type ApplicationChecker interface {
	HasApplication(ctx context.Context, activityID, memberID int64) (bool, error)
	DefaultApplication(ctx context.Context, activityID, memberID int64) error
}

func (c *ApplyModeCondition) Name() string { return "apply_mode" }

func (c *ApplyModeCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	// if c.Mode != types.ApplyManual {
	// 	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "mode=%d auto", c.Mode)
	// 	return pipeline.ResultPass
	// }
	if c.Repo == nil {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultError, "mode=%d repo=nil", c.Mode)
		return pipeline.ResultError
	}
	ok, err := c.Repo.HasApplication(ctx.Context, ctx.Activity.ID, ctx.Member.ID)
	if err != nil {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultError, "mode=%d has_application err=%v", c.Mode, err)
		return pipeline.ResultError
	}
	if !ok {
		if c.Mode == types.ApplyAuto {
			if err := c.Repo.DefaultApplication(ctx.Context, ctx.Activity.ID, ctx.Member.ID); err != nil {
				ctx.StepLog.Cond(c.Name(), pipeline.ResultError, "mode=%d default_application err=%v", c.Mode, err)
				return pipeline.ResultError
			}
			ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "mode=%d has_application=%v", c.Mode, ok)
			return pipeline.ResultPass
		}
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "mode=%d has_application=%v", c.Mode, ok)
		return pipeline.ResultSkip
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "mode=%d has_application=%v", c.Mode, ok)
	return pipeline.ResultPass
}
