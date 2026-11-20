package member

import (
	"context"
	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	"testing"
)

func TestContextMemberValidation(t *testing.T) {
	s := New(nil, nil, nil, nil)
	if _, err := s.GetMemberWithContextData(context.Background(), 1, ContextKindBasic); err == nil {
		t.Fatal("accepted missing tenant")
	}
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 1, UserType: ctxdata.UserTypeOperator, OperatorCode: "test_a"})
	for _, id := range []int64{0, -1} {
		if _, err := s.GetMemberWithContextData(ctx, id, ContextKindBasic); err == nil {
			t.Fatal("accepted invalid member id")
		}
	}
}

// Runs in the existing disposable PostgreSQL test schema.
func testContextQueries(t *testing.T, client *ent.Client, ctx, other context.Context, id int64, s *Service) {
	t.Helper()
	ctx = ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: ctxdata.OperatorCodeFromCtx(ctx)})
	other = ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: ctxdata.OperatorCodeFromCtx(other)})
	for _, kind := range []string{ContextKindBasic, ContextKindVerification} {
		m, err := s.GetMemberWithContextData(ctx, id, kind)
		if err != nil {
			t.Fatal(err)
		}
		if m.RegisterIP == nil || *m.RegisterIP != "192.0.2.1" || m.RegisterDeviceCode == nil || *m.RegisterDeviceCode != "test-device" {
			t.Fatal("registration source missing")
		}
		if kind == ContextKindBasic && m.AutoLevelID <= 0 {
			t.Fatal("missing auto level id")
		}
		if m.OperatorCode != "test_a" {
			t.Fatal("missing operator code")
		}
		if kind == ContextKindBasic && len(m.Edges.FixedLevelAssignments) != 0 {
			t.Fatal("expected no fixed levels")
		}
		if _, err = s.GetMemberWithContextData(other, id, kind); err == nil {
			t.Fatal("accepted foreign member")
		}
		if _, err = s.GetMemberWithContextData(ctx, 9223372036854775807, kind); err == nil {
			t.Fatal("accepted missing member")
		}
	}
}
