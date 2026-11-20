package betting_test

import (
	"context"
	"testing"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/betting"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type mockCrmServ struct {
	member     *types.MemberInfo
	opBaseInfo *types.OpBaseInfo
}

func (m *mockCrmServ) LoadMember(ctx context.Context, id int64) (*types.MemberInfo, error) {
	return m.member, nil
}

func (m *mockCrmServ) LoadOpBaseInfo(ctx context.Context) (*types.OpBaseInfo, error) {
	return m.opBaseInfo, nil
}

func TestService_Process(t *testing.T) {
	svc := betting.NewService(&mockCrmServ{
		member: &types.MemberInfo{ID: 1001, OperatorCode: "demo"},
	}, nil)
	evt := &betting.Event{
		MemberID: 1001,
		OrderID:  8001,
		Amount:   60000,
		GameID:   3,
	}
	out, user, err := svc.Process(context.Background(), evt)
	if err != nil {
		t.Fatal(err)
	}
	if out.OperatorCode != "demo" || out.ClientType != 2 {
		t.Fatalf("unexpected enriched event: %+v", out)
	}
	if user == nil || user.ID != 1001 {
		t.Fatalf("expected user 1001, got %+v", user)
	}
}

func TestService_Process_InvalidAmount(t *testing.T) {
	svc := betting.NewService(&mockCrmServ{member: &types.MemberInfo{ID: 1}}, nil)
	_, _, err := svc.Process(context.Background(), &betting.Event{
		MemberID: 1, OrderID: 1, Amount: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero bet amount")
	}
}

func TestMatchGameType(t *testing.T) {
	if !betting.MatchGameType([]int{1, 2, 3}, 2) {
		t.Fatal("expected game type 2 to match")
	}
	if betting.MatchGameType([]int{1, 2}, 5) {
		t.Fatal("expected game type 5 not to match")
	}
	if !betting.MatchGameType(nil, 99) {
		t.Fatal("empty allowed list should allow any game type")
	}
}
