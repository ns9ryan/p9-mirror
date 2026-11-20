package member

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
	"testing"
)

func TestContactUpdateError(t *testing.T) {
	for _, tt := range []struct {
		constraint string
		want       string
	}{
		{"uk_member_auth_operator_email_hash", "member.emailAlreadyBound"},
		{"uk_member_auth_operator_phone_hash", "member.phoneAlreadyBound"},
	} {
		t.Run(tt.constraint, func(t *testing.T) {
			err := fmt.Errorf("update contact: %w", &pgconn.PgError{Code: "23505", ConstraintName: tt.constraint})
			if got := contactUpdateError(err); got == nil || got.Error() != tt.want {
				t.Fatalf("got %v, want %s", got, tt.want)
			}
		})
	}
	for _, err := range []error{
		nil,
		errors.New("database unavailable"),
		&pgconn.PgError{Code: "23505", ConstraintName: "another_index"},
		&pgconn.PgError{Code: "23514", ConstraintName: "uk_member_auth_operator_email_hash"},
	} {
		if got := contactUpdateError(err); got != err {
			t.Fatalf("unexpected mapping: %v -> %v", err, got)
		}
	}
}

func TestContactRejectsInvalidInputBeforeDatabase(t *testing.T) {
	empty, badEmail, badPhone, longName := "", "invalid-email", "abc", ""
	for i := 0; i < 101; i++ {
		longName += "名"
	}
	zero := int64(0)
	for _, in := range []*memberpb.UpdateMemberContactRequest{
		nil, {MemberId: 1}, {MemberId: 1, Nickname: &empty}, {MemberId: 1, Nickname: &longName},
		{MemberId: 1, Phone: &badPhone}, {MemberId: 1, Email: &badEmail}, {MemberId: 1, AvatarId: &zero},
	} {
		if err := (&Service{}).UpdateContact(context.Background(), in, nil); err == nil {
			t.Fatalf("accepted invalid input: %v", in)
		}
	}
}
func TestContactRequiresEncryptionForSensitiveValues(t *testing.T) {
	value := "真实姓名"
	if err := (&Service{}).UpdateContact(context.Background(), &memberpb.UpdateMemberContactRequest{MemberId: 1, RealName: &value}, nil); err == nil {
		t.Fatal("accepted sensitive write without encryption")
	}
}
