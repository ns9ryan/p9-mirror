package referral_test

import (
	"context"
	"testing"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/referral"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type mockMemberServ struct {
	members map[int64]*types.MemberInfo
}

func (m *mockMemberServ) LoadMember(ctx context.Context, id int64) (*types.MemberInfo, error) {
	return m.members[id], nil
}

func TestService_ProcessReferralDeposit(t *testing.T) {
	svc := referral.NewService(&mockMemberServ{
		members: map[int64]*types.MemberInfo{
			1002: {ID: 1002, OperatorCode: "demo", InviterID: 1001},
			1001: {ID: 1001, OperatorCode: "demo"},
		},
	}, nil)
	evt := &referral.Event{
		EventType: referral.EventDeposit,
		MemberID:  1002,
		OrderID:   7001,
		Amount:    15000,
	}
	out, user, err := svc.Process(context.Background(), evt)
	if err != nil {
		t.Fatal(err)
	}
	if out.InviterID != 1001 {
		t.Fatalf("reward should belong to inviter 1001, got inviter %d", out.InviterID)
	}
	if user == nil || user.ID != 1001 {
		t.Fatalf("expected inviter user 1001, got %+v", user)
	}
}
