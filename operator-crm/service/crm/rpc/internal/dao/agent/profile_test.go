package agent

import (
	"context"
	"strings"
	"testing"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/i18n"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/agentpb"
)

func TestProfileRequiresOperator(t *testing.T) {
	s := New(nil)
	_, err := s.GetProfile(context.Background(), &agentpb.GetAgentProfileReq{AgentId: 1})
	assertApplyKey(t, err, i18n.Unauthorized)
}

func TestProfileRequiresAgentID(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.GetProfile(ctx, &agentpb.GetAgentProfileReq{})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestUpdateLoginPasswordRequiresSession(t *testing.T) {
	s := New(nil)
	_, err := s.UpdateLoginPassword(context.Background(), &agentpb.UpdateLoginPasswordReq{
		OldPassword: "old",
		NewPassword: "new",
	})
	assertApplyKey(t, err, i18n.Unauthorized)
}

func TestUpdateLoginPasswordRequiresPlaintext(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{
		UserID: 1, OperatorCode: "op01", Salt: "salt",
	})
	_, err := s.UpdateLoginPassword(ctx, &agentpb.UpdateLoginPasswordReq{OldPassword: " ", NewPassword: "new"})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestUpdateWithdrawPasswordRequiresPlaintext(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.UpdateWithdrawPassword(ctx, &agentpb.UpdateWithdrawPasswordReq{AgentId: 1, Password: "  "})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestTreeNodeAncestorAgentsRequiresKeyword(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.TreeNodeAncestorAgents(ctx, &agentpb.TreeNodeAncestorAgentsReq{Keyword: "  "})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestUpdateShareRateRange(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.UpdateShareRate(ctx, &agentpb.UpdateAgentShareRateReq{AgentId: 1, ShareRate: 101})
	assertApplyKey(t, err, i18n.AgentShareRateInvalid)
}

func TestUpdatePayoutRestrictedRange(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.UpdatePayoutRestricted(ctx, &agentpb.UpdateAgentPayoutRestrictedReq{AgentId: 1, PayoutRestricted: 4})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestUpdateStatusRange(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.UpdateStatus(ctx, &agentpb.UpdateAgentStatusReq{AgentId: 1, Status: 4})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestUpdateProfileFieldType(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.UpdateProfile(ctx, &agentpb.UpdateAgentProfileReq{AgentId: 1, FieldType: 13})
	assertApplyKey(t, err, i18n.InvalidParam)
}

func TestUpdateRemarkLength(t *testing.T) {
	s := New(nil)
	ctx := ctxdata.WithClaims(context.Background(), &ctxdata.Claims{OperatorCode: "op01"})
	_, err := s.UpdateRemark(ctx, &agentpb.UpdateAgentRemarkReq{AgentId: 1, Remark: strings.Repeat("a", 1001)})
	assertApplyKey(t, err, i18n.InvalidParam)
}
