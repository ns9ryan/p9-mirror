package deposit_test

import (
	"context"
	"testing"

	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/deposit"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

type mockCrmServ struct {
	member *types.MemberInfo
}

func (m *mockCrmServ) LoadMember(ctx context.Context, id int64) (*types.MemberInfo, error) {
	return m.member, nil
}

func (m *mockCrmServ) LoadOpBaseInfo(ctx context.Context) (*types.OpBaseInfo, error) {
	return &types.OpBaseInfo{Currency: "USD", TimeZone: "UTC"}, nil
}

func TestService_Process(t *testing.T) {
	svc := deposit.NewService(&mockCrmServ{
		member: &types.MemberInfo{ID: 1001, OperatorCode: "demo"},
	}, nil)
	evt := &deposit.Event{
		MemberID: 1001,
		OrderID:  9001,
		Amount:   20000,
	}
	out, user, err := svc.Process(context.Background(), evt)
	if err != nil {
		t.Fatal(err)
	}
	if out.ClientType != 2 {
		t.Fatalf("expected client_type=2, got %d", out.ClientType)
	}
	if user == nil || user.ID != 1001 {
		t.Fatalf("expected user 1001, got %+v", user)
	}
}

func TestService_Process_InvalidAmount(t *testing.T) {
	svc := deposit.NewService(&mockCrmServ{member: &types.MemberInfo{ID: 1}}, nil)
	_, _, err := svc.Process(context.Background(), &deposit.Event{MemberID: 1, OrderID: 1, Amount: 0})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}
