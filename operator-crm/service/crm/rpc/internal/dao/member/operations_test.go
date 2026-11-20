package member

import (
	"context"
	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
	"testing"
)

func testMemberOperations(t *testing.T, c *ent.Client, ctx, other context.Context, id int64, s *Service) {
	t.Helper()
	auth := c.MemberAuth.Create().SetMemberID(id).SetSmsCodeErrorCount(4).SetEmailCodeErrorCount(5).SetLoginPasswordErrorCount(6).SetWithdrawalPasswordErrorCount(7).SaveX(ctx)
	profile := c.MemberProfile.Create().SetMemberID(id).SetNickname("test").SetAvatarID(0).SetAvatarFrameID(0).SaveX(ctx)

	cryptor, err := fieldcrypto.New([]byte("0123456789abcdef"), []byte("test-hmac-key"))
	if err != nil {
		t.Fatal(err)
	}
	phone, email, name := "+8613800000000", "MEMBER@example.com", "张三"
	request := &memberpb.UpdateMemberContactRequest{MemberId: id, Phone: &phone, Email: &email, RealName: &name}
	if err := s.UpdateContact(ctx, request, cryptor); err == nil {
		t.Fatal("contact change without proof succeeded")
	}
	for _, input := range []struct {
		channel int32
		target  string
		proof   **memberpb.ContactBindingCodeProof
	}{{ContactBindingCodeSMS, phone, &request.PhoneVerification}, {ContactBindingCodeEmail, email, &request.EmailVerification}} {
		pending, err := s.CreateContactBindingCode(ctx, id, input.channel, input.target, cryptor)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.FinishContactBindingCodeDelivery(ctx, pending.RecordID, true); err != nil {
			t.Fatal(err)
		}
		*input.proof = &memberpb.ContactBindingCodeProof{RecordId: pending.RecordID, VerificationCode: pending.Code}
	}
	filters := ContactBindingCodeFilters{MemberID: id}
	records, total, err := s.ListContactBindingCodes(ctx, "test_a", filters, 0, 20)
	if err != nil || total != 2 || len(records) != 2 {
		t.Fatalf("list active codes: total=%d, error=%v", total, err)
	}
	for _, record := range records {
		if record.VerificationCode != "" {
			t.Fatal("list disclosed an active verification code")
		}
	}
	correctEmailCode := request.EmailVerification.VerificationCode
	request.EmailVerification.VerificationCode = "000000"
	if correctEmailCode == "000000" {
		request.EmailVerification.VerificationCode = "000001"
	}
	if err := s.UpdateContact(ctx, request, cryptor); err == nil {
		t.Fatal("wrong email code accepted")
	}
	if got := c.MemberAuth.GetX(ctx, auth.ID); got.PhoneCiphertext != nil || got.EmailCiphertext != nil {
		t.Fatal("failed verification changed contact fields")
	}
	request.EmailVerification.VerificationCode = correctEmailCode
	err = s.UpdateContact(ctx, request, cryptor)
	if err != nil {
		t.Fatal(err)
	}
	records, total, err = s.ListContactBindingCodes(ctx, "test_a", filters, 0, 20)
	if err != nil || total != 2 || len(records) != 2 {
		t.Fatalf("list consumed codes: total=%d, error=%v", total, err)
	}
	expectedCodes := map[int32]string{ContactBindingCodeSMS: request.PhoneVerification.VerificationCode, ContactBindingCodeEmail: correctEmailCode}
	for _, record := range records {
		if record.Status != 2 || record.UsedAt == nil || record.VerificationCode != expectedCodes[record.Channel] {
			t.Fatal("verified code was not consumed")
		}
		if record.Channel == ContactBindingCodeEmail && record.FailedAttempts != 1 {
			t.Fatal("wrong code attempt was not recorded")
		}
	}
	gotAuth := c.MemberAuth.GetX(ctx, auth.ID)
	plain, err := cryptor.Open(*gotAuth.PhoneCiphertext)
	if err != nil || plain != "8613800000000" {
		t.Fatal("phone encryption/normalization failed", err)
	}
	hash, _ := cryptor.LookupHash("member@example.com")
	if gotAuth.EmailLookupHash == nil || *gotAuth.EmailLookupHash != hash {
		t.Fatal("email lookup hash mismatch")
	}
	gotProfile := c.MemberProfile.GetX(ctx, profile.ID)
	plain, err = cryptor.Open(*gotProfile.RealNameCiphertext)
	if err != nil || plain != name || gotProfile.Nickname != "test" {
		t.Fatal("partial profile update failed")
	}
	if err = s.UpdateContact(other, &memberpb.UpdateMemberContactRequest{MemberId: id, RealName: &name}, cryptor); err == nil {
		t.Fatal("cross-operator update accepted")
	}
	empty := ""
	if err := s.UpdateContact(ctx, &memberpb.UpdateMemberContactRequest{MemberId: id, Phone: &empty}, nil); err == nil {
		t.Fatal("clearing phone without proof succeeded")
	}
	if err := s.UpdateContact(ctx, request, cryptor); err == nil {
		t.Fatal("consumed verification codes were accepted")
	}
	if err = s.UpdateContact(ctx, &memberpb.UpdateMemberContactRequest{MemberId: id}, cryptor); err == nil {
		t.Fatal("empty update accepted")
	}
	for _, status := range []int64{2, 3, 1} {
		if err := s.UpdateStatus(ctx, id, status); err != nil {
			t.Fatal(err)
		}
		if c.Member.GetX(ctx, id).Status != status {
			t.Fatal("status not updated")
		}
	}
	if err := s.UpdateStatus(ctx, id, 4); err == nil {
		t.Fatal("accepted invalid status")
	}
	if err := s.UpdateStatus(other, id, 3); err == nil {
		t.Fatal("accepted foreign member")
	}
	for _, field := range []string{"sms_code_error_count", "email_code_error_count", "login_password_error_count", "withdrawal_password_error_count"} {
		c.MemberAuth.UpdateOneID(auth.ID).SetSmsCodeErrorCount(4).SetEmailCodeErrorCount(5).SetLoginPasswordErrorCount(6).SetWithdrawalPasswordErrorCount(7).SaveX(ctx)
		if err := s.ResetAuthErrors(ctx, id, field); err != nil {
			t.Fatal(err)
		}
		got := c.MemberAuth.GetX(ctx, auth.ID)
		values := map[string]int64{"sms_code_error_count": got.SmsCodeErrorCount, "email_code_error_count": got.EmailCodeErrorCount, "login_password_error_count": got.LoginPasswordErrorCount, "withdrawal_password_error_count": got.WithdrawalPasswordErrorCount}
		want := map[string]int64{"sms_code_error_count": 4, "email_code_error_count": 5, "login_password_error_count": 6, "withdrawal_password_error_count": 7}
		want[field] = 0
		for k, v := range want {
			if values[k] != v {
				t.Fatalf("reset %s changed %s to %d", field, k, values[k])
			}
		}
	}
	if err := s.ResetAuthErrors(ctx, id, "login_password_hash"); err == nil {
		t.Fatal("accepted invalid reset field")
	}
	for _, remark := range []string{"重点关注", ""} {
		if err := s.UpdateRemark(ctx, id, remark); err != nil {
			t.Fatal(err)
		}
		p := c.MemberProfile.GetX(ctx, profile.ID)
		if p.Remark == nil || *p.Remark != remark {
			t.Fatal("remark mismatch")
		}
	}
	for _, disabled := range []bool{true, false} {
		if err := s.SetChatDisabled(ctx, id, disabled); err != nil {
			t.Fatal(err)
		}
		m := c.Member.GetX(ctx, id)
		if m.ChatDisabled != disabled || m.BettingDisabled || m.WithdrawalDisabled {
			t.Fatal("chat changed other flags")
		}
	}
	for _, disabled := range []bool{true, false} {
		if err := s.SetBettingDisabled(ctx, id, disabled); err != nil {
			t.Fatal(err)
		}
		if c.Member.GetX(ctx, id).BettingDisabled != disabled {
			t.Fatal("betting mismatch")
		}
	}
	for _, disabled := range []bool{true, false} {
		if err := s.SetWithdrawalDisabled(ctx, id, disabled); err != nil {
			t.Fatal(err)
		}
		if c.Member.GetX(ctx, id).WithdrawalDisabled != disabled {
			t.Fatal("withdrawal mismatch")
		}
	}
	if err := s.UpdateRemark(other, id, "wrong"); err == nil {
		t.Fatal("accepted foreign remark")
	}
	if err := s.ResetAuthErrors(other, id, "sms_code_error_count"); err == nil {
		t.Fatal("accepted foreign auth reset")
	}
	if err := s.SetChatDisabled(other, id, true); err == nil {
		t.Fatal("accepted foreign restriction")
	}
	c.MemberAuth.DeleteOneID(auth.ID).ExecX(ctx)
	if err := s.ResetAuthErrors(ctx, id, "sms_code_error_count"); err == nil {
		t.Fatal("accepted missing auth row")
	}
	c.MemberProfile.DeleteOneID(profile.ID).ExecX(ctx)
	if err := s.UpdateRemark(ctx, id, "missing"); err == nil {
		t.Fatal("accepted missing profile row")
	}
}
