package condition

import (
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type ActivityStatusCondition struct{}

func (c *ActivityStatusCondition) Name() string { return "activity_status" }

func (c *ActivityStatusCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	status := ctx.Activity.Status
	want := types.ActivityStatusOngoing
	if status != want {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "status=%d want=%d", status, want)
		return pipeline.ResultSkip
	}
	ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "status=%d", status)
	return pipeline.ResultPass
}

type ActivityTimeCondition struct{}

func (c *ActivityTimeCondition) Name() string { return "activity_time" }

func (c *ActivityTimeCondition) Evaluate(ctx *pipeline.EvalContext) pipeline.Result {
	et := ctx.Event.GetEventTime()
	openAt, closeAt := ctx.Activity.OpenAt, ctx.Activity.CloseAt
	// 活动未发布
	if !ctx.Activity.IsRelease() {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "promo_not_released event_time=%s",
			et.Format("2006-01-02 15:04:05"))
		return pipeline.ResultSkip
	}
	// 活动是否已经结束
	if ctx.Activity.IsEnd(et) {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultSkip, "event_time=%s open_at=%s close_at=%s",
			et.Format("2006-01-02 15:04:05"), openAt.Format("2006-01-02 15:04:05"), closeAt.Format("2006-01-02 15:04:05"))
		return pipeline.ResultSkip
	}
	// 活动永久
	if ctx.Activity.IsPermanent {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "event_time=%s permanent=true",
			et.Format("2006-01-02 15:04:05"))
	} else {
		ctx.StepLog.Cond(c.Name(), pipeline.ResultPass, "event_time=%s open_at=%s close_at=%s",
			et.Format("2006-01-02 15:04:05"),
			openAt.Format("2006-01-02 15:04:05"),
			closeAt.Format("2006-01-02 15:04:05"))
	}
	return pipeline.ResultPass
}
