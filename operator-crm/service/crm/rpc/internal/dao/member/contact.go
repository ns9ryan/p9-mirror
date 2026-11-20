package member

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/modules/fieldcrypto"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberprofile"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"
)

// UpdateContact 在当前运营商范围内原子更新请求中提供的会员资料。
func (s *Service) UpdateContact(ctx context.Context, in *memberpb.UpdateMemberContactRequest, cryptor *fieldcrypto.Cryptor) error {
	bad := func() error { return xerr.BadRequest("common.invalidParam") }
	if in == nil || (in.AvatarId == nil && in.AvatarFrameId == nil && in.Nickname == nil && in.RealName == nil && in.Phone == nil && in.IdentityNumber == nil && in.Email == nil) {
		return bad()
	}
	if (in.AvatarId != nil && *in.AvatarId <= 0) || (in.AvatarFrameId != nil && *in.AvatarFrameId <= 0) {
		return bad()
	}
	for _, f := range []struct {
		v   *string
		max int
	}{{in.Nickname, 100}, {in.RealName, 100}, {in.IdentityNumber, 100}, {in.Phone, 32}, {in.Email, 254}} {
		if f.v != nil && (!utf8.ValidString(*f.v) || utf8.RuneCountInString(*f.v) > f.max) {
			return bad()
		}
	}
	if in.Nickname != nil && strings.TrimSpace(*in.Nickname) == "" {
		return bad()
	}
	contacts := []struct {
		channel int32
		value   *string
		proof   *memberpb.ContactBindingCodeProof
		target  string
	}{
		{channel: ContactBindingCodeSMS, value: in.Phone, proof: in.PhoneVerification},
		{channel: ContactBindingCodeEmail, value: in.Email, proof: in.EmailVerification},
	}
	for index := range contacts {
		contact := &contacts[index]
		if contact.value == nil {
			if contact.proof != nil {
				return bad()
			}
			continue
		}
		target, err := normalizeContactBindingTarget(contact.channel, *contact.value)
		if err != nil {
			return err
		}
		if contact.proof.GetRecordId() <= 0 || !contactVerificationCode.MatchString(contact.proof.GetVerificationCode()) {
			return xerr.BadRequest("member.contactBindingCodeInvalid")
		}
		contact.target = target
	}

	// 敏感字段在事务开始前加密，手机号和邮箱同时生成用于查找的哈希。
	seal := func(v *string, normalize func(string) string) (*string, *string, error) {
		if v == nil {
			return nil, nil, nil
		}
		value := strings.TrimSpace(*v)
		if normalize != nil {
			value = normalize(value)
		}
		if value == "" {
			return nil, nil, nil
		}
		if cryptor == nil {
			return nil, nil, fmt.Errorf("field encryption is not configured")
		}
		ct, e := cryptor.Seal(value)
		if e != nil {
			return nil, nil, e
		}
		var hash *string
		if normalize != nil {
			h, e := cryptor.LookupHash(value)
			if e != nil {
				return nil, nil, e
			}
			hash = &h
		}
		return &ct, hash, nil
	}
	real, _, err := seal(in.RealName, nil)
	if err != nil {
		return err
	}
	identity, _, err := seal(in.IdentityNumber, nil)
	if err != nil {
		return err
	}
	phone, ph, err := seal(in.Phone, fieldcrypto.NormalizePhone)
	if err != nil {
		return err
	}
	email, eh, err := seal(in.Email, fieldcrypto.NormalizeEmail)
	if err != nil {
		return err
	}
	var invalidVerification bool
	err = s.withMemberUpdateTx(ctx, in.MemberId, func(tx *ent.Tx) error {
		driver := tx.DialectDriver()
		var verifiedRecordIDs []int64
		for _, contact := range contacts {
			if contact.value == nil {
				continue
			}
			valid, err := verifyContactBindingCode(ctx, driver, in.MemberId, contact.channel, contact.target, contact.proof, cryptor)
			if err != nil {
				return err
			}
			if !valid {
				// Commit the failed attempt without changing the member or consuming either code.
				invalidVerification = true
				return nil
			}
			verifiedRecordIDs = append(verifiedRecordIDs, contact.proof.GetRecordId())
		}
		for _, recordID := range verifiedRecordIDs {
			if err := driver.Exec(ctx, `UPDATE member_contact_binding_code
                SET status = 2, used_at = clock_timestamp()
                WHERE operator_code = $1 AND member_id = $2 AND id = $3`,
				[]any{ctxdata.OperatorCodeFromCtx(ctx), in.MemberId, recordID}, nil); err != nil {
				return err
			}
		}
		p, e := tx.MemberProfile.Query().Where(memberprofile.MemberIDEQ(in.MemberId)).Only(ctx)
		if ent.IsNotFound(e) {
			return bad()
		}
		if e != nil {
			return e
		}
		// 未提供的字段保留原值，显式传入空白的敏感字段会清除原值。
		u := tx.MemberProfile.UpdateOneID(p.ID)
		if in.AvatarId != nil {
			u.SetAvatarID(*in.AvatarId)
		}
		if in.AvatarFrameId != nil {
			u.SetAvatarFrameID(*in.AvatarFrameId)
		}
		if in.Nickname != nil {
			u.SetNickname(strings.TrimSpace(*in.Nickname))
		}
		if in.RealName != nil {
			if real == nil {
				u.ClearRealNameCiphertext()
			} else {
				u.SetRealNameCiphertext(*real)
			}
		}
		if in.IdentityNumber != nil {
			if identity == nil {
				u.ClearIdentityNumberCiphertext()
			} else {
				u.SetIdentityNumberCiphertext(*identity)
			}
		}
		if e = u.Exec(ctx); e != nil {
			return e
		}
		// 在同一事务内更新认证信息，密文和查找哈希一并写入。
		if in.Phone != nil || in.Email != nil {
			a, e := tx.MemberAuth.Query().Where(memberauth.MemberIDEQ(in.MemberId)).Only(ctx)
			if ent.IsNotFound(e) {
				return bad()
			}
			if e != nil {
				return e
			}
			u := tx.MemberAuth.UpdateOneID(a.ID)
			if in.Phone != nil {
				u.SetPhoneCiphertext(*phone).SetPhoneLookupHash(*ph)
			}
			if in.Email != nil {
				u.SetEmailCiphertext(*email).SetEmailLookupHash(*eh)
			}
			return u.Exec(ctx)
		}
		return nil
	})
	if err == nil && invalidVerification {
		return xerr.BadRequest("member.contactBindingCodeInvalid")
	}
	return contactUpdateError(err)
}

// Keep the unique index as the authority, including concurrent binding requests.
func contactUpdateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "uk_member_auth_operator_email_hash":
			return xerr.BadRequest("member.emailAlreadyBound")
		case "uk_member_auth_operator_phone_hash":
			return xerr.BadRequest("member.phoneAlreadyBound")
		}
	}
	if ent.IsConstraintError(err) {
		return xerr.BadRequest("common.invalidParam")
	}
	return err
}

// 获取会员联系资料
func (s *Service) GetMemberContact(ctx context.Context, memberID int64) (*ent.MemberProfile, *ent.MemberAuth, error, error) {
	beforeProfile, beforeProfileErr := s.db.MemberProfile.Query().
		Where(memberprofile.MemberIDEQ(memberID)).
		Only(ctx)

	beforeAuth, beforeAuthErr := s.db.MemberAuth.Query().
		Where(memberauth.MemberIDEQ(memberID)).
		Only(ctx)

	return beforeProfile, beforeAuth, beforeProfileErr, beforeAuthErr
}
