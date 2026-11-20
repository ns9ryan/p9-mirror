package reward_test

import (
	"context"
	"testing"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/reward"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

func TestFixedCalculator(t *testing.T) {
	calc := &reward.FixedCalculator{
		BaseCalculator: reward.BaseCalculator{
			Schemas: []types.RewardSchema{
				{ID: 1, Threshold: 1000, Amount: 500},
			},
		},
	}
	ctx := &pipeline.EvalContext{
		Context:  context.Background(),
		Activity: &types.Activity{ID: 1},
		Member:   &types.MemberInfo{ID: 100},
		Event:    stubEvent{orderID: 999},
		Metric:   &types.TriggerMetric{DepositAmount: 2000},
		StepLog:  pipeline.NewStepLogger(1, 100),
	}
	items, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Amount != 500 {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestPercentCalculator_WithCap(t *testing.T) {
	calc := &reward.PercentCalculator{
		BaseCalculator: reward.BaseCalculator{
			Schemas: []types.RewardSchema{
				{ID: 1, Threshold: 1000, Rate: 1000, MaxBonus: 500},
			},
		},
	}
	ctx := &pipeline.EvalContext{
		Context:  context.Background(),
		Activity: &types.Activity{ID: 1},
		Member:   &types.MemberInfo{ID: 100},
		Event:    stubEvent{amount: 10000, currency: "CNY"},
		Metric:   &types.TriggerMetric{DepositAmount: 10000},
		StepLog:  pipeline.NewStepLogger(1, 100),
	}
	items, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Amount != 500 {
		t.Fatalf("expected capped amount 500, got %+v", items)
	}
}

func TestFixedCalculator_ReferralCount(t *testing.T) {
	calc := &reward.FixedCalculator{
		BaseCalculator: reward.BaseCalculator{
			Schemas: []types.RewardSchema{
				{ID: 1, Count: 3, Amount: 10000},
			},
		},
	}
	ctx := &pipeline.EvalContext{
		Context:  context.Background(),
		Activity: &types.Activity{ID: 3},
		Member:   &types.MemberInfo{ID: 1001},
		Event:    stubEvent{},
		Metric:   &types.TriggerMetric{ValidReferralCount: 3},
		StepLog:  pipeline.NewStepLogger(3, 1001),
	}
	items, err := calc.Calculate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Amount != 10000 {
		t.Fatalf("unexpected referral reward: %+v", items)
	}
}

type stubEvent struct {
	orderID  int64
	amount   int64
	currency string
	timeZone string
}

func (s stubEvent) GetOperatorCode() string       { return "" }
func (s stubEvent) GetMemberID() int64            { return 0 }
func (s stubEvent) GetOrderID() int64             { return s.orderID }
func (s stubEvent) GetAmount() int64              { return s.amount }
func (s stubEvent) GetCurrency() string           { return s.currency }
func (s stubEvent) GetTimeZone() string           { return s.timeZone }
func (s stubEvent) GetClientType() int            { return 0 }
func (s stubEvent) GetGameID() int64              { return 0 }
func (s stubEvent) GetEventTime() time.Time       { return time.Time{} }
func (s stubEvent) GetEventType() types.EventType { return 0 }
func (s stubEvent) GetDepositCount() int          { return 0 }
