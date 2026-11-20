package audit_test

import (
	"testing"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/audit"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

func TestConfigStrategy_Decide(t *testing.T) {
	cases := []struct {
		name   string
		auto   bool
		maxAmt int64
		amount int64
		want   audit.Decision
	}{
		{"auto_off", false, 1000, 100, audit.DecisionPending},
		{"auto_below", true, 1000, 999, audit.DecisionAutoPass},
		{"auto_equal", true, 1000, 1000, audit.DecisionPending},
		{"auto_above", true, 1000, 1001, audit.DecisionPending},
		{"auto_zero_threshold", true, 0, 1, audit.DecisionPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &audit.ConfigStrategy{AutoAudit: tc.auto, AutoAuditMaxAmt: tc.maxAmt}
			got := s.Decide(&types.ActivityOrder{RewardsTotal: tc.amount})
			if got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
