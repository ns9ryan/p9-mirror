package agent

import (
	"context"
	"errors"
	"testing"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

func TestAddAgentApplyRequiresOperator(t *testing.T) {
	s := New(nil)
	_, err := s.AddAgentApply(context.Background(), &agentpb.AddAgentApplyReq{
		Account: "a", Realname: "n", Mobile: "1", RegisterSource: 1, ApplyShareRate: 100,
	})
	assertApplyKey(t, err, i18n.Unauthorized)
}

func TestAddAgentApplyRequiresFields(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.AddAgentApply(ctx, &agentpb.AddAgentApplyReq{Realname: "n", Mobile: "1", RegisterSource: 1})
	assertApplyKey(t, err, i18n.AgentFieldsRequired)
}

func TestReviewAgentApplyRequiresID(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01", Username: "admin"})
	_, err := s.ReviewAgentApply(ctx, &agentpb.ReviewAgentApplyReq{Id: 0, Action: reviewActionManualReject})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestReviewAuditor(t *testing.T) {
	if got := reviewAuditor(context.Background()); got != "" {
		t.Fatalf("empty ctx auditor = %q", got)
	}
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{Username: "alice", UserID: 7})
	if got := reviewAuditor(ctx); got != "alice" {
		t.Fatalf("auditor = %q, want alice", got)
	}
	ctx = ctxdata.WithClaims(context.Background(), &ctxdata.Claims{UserID: 9})
	if got := reviewAuditor(ctx); got != "9" {
		t.Fatalf("auditor = %q, want 9", got)
	}
}

func assertApplyKey(t *testing.T, err error, key string) {
	t.Helper()
	var e *xerr.Error
	if !errors.As(err, &e) || e.Message != key {
		t.Fatalf("err = %v, want %s", err, key)
	}
}
