package condition_test

import (
	"context"
	"testing"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/condition"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/metric"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

func TestActivityTimeCondition_Pass(t *testing.T) {
	now := time.Now()
	openAt := now.Add(-time.Hour)
	closeAt := now.Add(time.Hour)
	ctx := &pipeline.EvalContext{
		Context: context.Background(),
		Activity: &types.Activity{
			Status:    types.ActivityStatusOngoing,
			OpenAt:    &openAt,
			CloseAt:   &closeAt,
			ReleaseAt: &now,
		},
		Event:   stubEvent{eventTime: now},
		StepLog: pipeline.NewStepLogger(1, 1),
	}
	chain := condition.NewChain(
		&condition.ActivityStatusCondition{},
		&condition.ActivityTimeCondition{},
	)
	if got := chain.Run(ctx); got != pipeline.ResultPass {
		t.Fatalf("expected pass, got %d", got)
	}
}

func TestVipLevelCondition_Skip(t *testing.T) {
	ctx := &pipeline.EvalContext{
		Context:  context.Background(),
		Activity: &types.Activity{},
		Member:   &types.MemberInfo{AutoLevel: 1},
		Event:    stubEvent{},
		StepLog:  pipeline.NewStepLogger(1, 1),
	}
	cond := &condition.AutoLevelCondition{Allowed: []int64{2, 3}}
	if got := cond.Evaluate(ctx); got != pipeline.ResultSkip {
		t.Fatalf("expected skip, got %d", got)
	}
}

func TestCycleModeCondition_Daily(t *testing.T) {
	now := time.Now()
	openAt := now.Add(-24 * time.Hour)
	ctx := &pipeline.EvalContext{
		Context: context.Background(),
		Activity: &types.Activity{
			OpenAt:    &openAt,
			CycleMode: types.CycleDaily,
		},
		Event:   stubEvent{eventTime: now},
		StepLog: pipeline.NewStepLogger(1, 1),
	}
	cond := &condition.CycleModeCondition{Manager: pipeline.NewManager()}
	if got := cond.Evaluate(ctx); got != pipeline.ResultPass {
		t.Fatalf("expected pass, got %d", got)
	}
	if ctx.Round < 1 {
		t.Fatalf("expected round >= 1, got %d", ctx.Round)
	}
}

func TestDepositTypeCondition_FirstDeposit(t *testing.T) {
	ctx := &pipeline.EvalContext{
		Context: context.Background(),
		Activity: &types.Activity{
			Settings: types.ActivitySettings{DepositRewardType: types.DepositFirst},
		},
		Event:   stubEvent{depositCount: 1, amount: 10000},
		StepLog: pipeline.NewStepLogger(1, 1),
	}
	cond := &condition.DepositTypeCondition{
		RewardType: types.DepositFirst,
		Calc:       &metric.DepositCalculator{},
	}
	if got := cond.Evaluate(ctx); got != pipeline.ResultPass {
		t.Fatalf("expected pass, got %d", got)
	}
	if ctx.Metric == nil || ctx.Metric.DepositCount != 1 || ctx.Metric.DepositAmount != 10000 {
		t.Fatalf("metric not loaded by condition: %+v", ctx.Metric)
	}
}

type stubEvent struct {
	eventTime    time.Time
	currency     string
	timeZone     string
	amount       int64
	depositCount int
}

func (s stubEvent) GetOperatorCode() string       { return "" }
func (s stubEvent) GetMemberID() int64            { return 0 }
func (s stubEvent) GetOrderID() int64             { return 0 }
func (s stubEvent) GetAmount() int64              { return s.amount }
func (s stubEvent) GetCurrency() string           { return s.currency }
func (s stubEvent) GetTimeZone() string           { return s.timeZone }
func (s stubEvent) GetClientType() int            { return 0 }
func (s stubEvent) GetGameID() int64              { return 0 }
func (s stubEvent) GetEventTime() time.Time       { return s.eventTime }
func (s stubEvent) GetEventType() types.EventType { return 0 }
func (s stubEvent) GetDepositCount() int          { return s.depositCount }
