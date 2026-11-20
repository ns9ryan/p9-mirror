package condition

import "oa.98ent.com/p9/operator-activity/modules/activity/pipeline"

// Condition 责任链节点
type Condition interface {
	Name() string
	Evaluate(ctx *pipeline.EvalContext) pipeline.Result
}

// Chain 条件责任链
type Chain struct {
	conditions []Condition
}

func NewChain(conditions ...Condition) *Chain {
	return &Chain{conditions: conditions}
}

func (c *Chain) Add(cond Condition) *Chain {
	c.conditions = append(c.conditions, cond)
	return c
}

func (c *Chain) Run(ctx *pipeline.EvalContext) pipeline.Result {
	for _, cond := range c.conditions {
		result := cond.Evaluate(ctx)
		switch result {
		case pipeline.ResultSkip:
			return pipeline.ResultSkip
		case pipeline.ResultError:
			if ctx.StepLog != nil {
				ctx.StepLog.Cond(cond.Name(), pipeline.ResultError, "evaluate error")
			}
			return pipeline.ResultError
		}
	}
	return pipeline.ResultPass
}
