package condition

import (
	"context"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// LimitChecker 限制查询接口
type LimitChecker interface {
	CountOrders(ctx context.Context, req LimitCountRequest) (int, error)
	CountApplications(ctx context.Context, req LimitCountRequest) (int, error)
}

type LimitCountRequest struct {
	ActivityID int64
	MemberID   int64
	Round      int
	Window     types.TimeWindow
	Scope      string
}

type ApplyLimitCondition struct {
	Limits []types.ApplyLimitType
	Repo   LimitChecker
}

func (c *ApplyLimitCondition) Name() string { return "apply_limit" }

func (c *ApplyLimitCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	if len(c.Limits) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "no_limit")
		return pipeline.ResultPass
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "limits=%v", c.Limits)
	return pipeline.ResultPass
}

type ClaimLimitCondition struct {
	Limits []types.ClaimLimitType
	Repo   LimitChecker
}

func (c *ClaimLimitCondition) Name() string { return "claim_limit" }

func (c *ClaimLimitCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	if len(c.Limits) == 0 {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "no_limit")
		return pipeline.ResultPass
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "limits=%v", c.Limits)
	return pipeline.ResultPass
}

func checkLimit333(ctx *pipeline.EvalContext, name string, limit types.LimitConfig, repo LimitChecker, isApply bool) pipeline.Result {
	if repo == nil {
		ctx.StepLog.Cond(name, pipeline.ResultPass, "repo=nil skip_check")
		return pipeline.ResultPass
	}
	req := LimitCountRequest{
		ActivityID: ctx.Activity.ID,
		MemberID:   ctx.Member.ID,
		Round:      ctx.Round,
		Window:     ctx.Window,
	}
	checks := []struct {
		limit int
		scope string
	}{
		{limit.PerUser, "per_user"},
		{limit.PerDay, "per_day"},
		{limit.PerWeek, "per_week"},
		{limit.PerMonth, "per_month"},
		{limit.PerCycle, "per_cycle"},
		{limit.Total, "total"},
	}
	var details []string
	for _, chk := range checks {
		if chk.limit <= 0 {
			continue
		}
		req.Scope = chk.scope
		var count int
		var err error
		if isApply {
			count, err = repo.CountApplications(ctx.Context, req)
		} else {
			count, err = repo.CountOrders(ctx.Context, req)
		}
		if err != nil {
			ctx.StepLog.Cond(name, pipeline.ResultError, "scope=%s limit=%d err=%v", chk.scope, chk.limit, err)
			return pipeline.ResultError
		}
		details = append(details, chk.scope+"="+itoa(count)+"/"+itoa(chk.limit))
		if count >= chk.limit {
			ctx.StepLog.Cond(name, pipeline.ResultSkip, "scope=%s count=%d limit=%d exceeded",
				chk.scope, count, chk.limit)
			return pipeline.ResultSkip
		}
	}
	if len(details) == 0 {
		ctx.StepLog.Cond(name, pipeline.ResultPass, "no_limit_configured")
		return pipeline.ResultPass
	}
	ctx.StepLog.Cond(name, pipeline.ResultPass, "checks=%v", details)
	return pipeline.ResultPass
}

type IdempotencyCondition struct {
	Repo IdempotencyChecker
}

type IdempotencyChecker interface {
	Exists(key string) (bool, error)
}

func (c *IdempotencyCondition) Name() string { return "idempotency" }

func (c *IdempotencyCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "deferred_to_engine")
	return pipeline.ResultPass
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
