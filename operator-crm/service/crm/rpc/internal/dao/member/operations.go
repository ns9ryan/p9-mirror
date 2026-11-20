package member

import (
	"context"
	"unicode/utf8"

	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberprofile"
)

const (
	statusNormal   int64 = 1
	statusFrozen   int64 = 2
	statusDisabled int64 = 3
)

// IsValidStatus 判断状态是否为正常、冻结或禁用。
func IsValidStatus(status int64) bool {
	switch status {
	case statusNormal, statusFrozen, statusDisabled:
		return true
	default:
		return false
	}
}

const maxRemarkRunes = 200

func (s *Service) UpdateStatus(ctx context.Context, id, status int64) error {
	if !IsValidStatus(status) {
		return xerr.BadRequest("common.invalidParam")
	}
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error { return tx.Member.UpdateOneID(id).SetStatus(status).Exec(ctx) })
}

// ResetAuthErrors 将指定的认证错误计数清零，字段必须在允许列表中。
func (s *Service) ResetAuthErrors(ctx context.Context, id int64, field string) error {
	switch field {
	case memberauth.FieldSmsCodeErrorCount, memberauth.FieldEmailCodeErrorCount, memberauth.FieldLoginPasswordErrorCount, memberauth.FieldWithdrawalPasswordErrorCount:
	default:
		return xerr.BadRequest("common.invalidParam")
	}
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error {
		u := tx.MemberAuth.Update().Where(memberauth.MemberIDEQ(id))
		switch field {
		case memberauth.FieldSmsCodeErrorCount:
			u.SetSmsCodeErrorCount(0)
		case memberauth.FieldEmailCodeErrorCount:
			u.SetEmailCodeErrorCount(0)
		case memberauth.FieldLoginPasswordErrorCount:
			u.SetLoginPasswordErrorCount(0)
		case memberauth.FieldWithdrawalPasswordErrorCount:
			u.SetWithdrawalPasswordErrorCount(0)
		}
		n, err := u.Save(ctx)
		if err != nil {
			return err
		}
		if n != 1 {
			return xerr.BadRequest("common.invalidParam")
		}
		return nil
	})
}

// UpdateRemark 按 Unicode 字符数限制备注长度，并要求更新恰好一条会员资料。
func (s *Service) UpdateRemark(ctx context.Context, id int64, remark string) error {
	if !utf8.ValidString(remark) || utf8.RuneCountInString(remark) > maxRemarkRunes {
		return xerr.BadRequest("common.invalidParam")
	}
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error {
		n, err := tx.MemberProfile.Update().Where(memberprofile.MemberIDEQ(id)).SetRemark(remark).Save(ctx)
		if err != nil {
			return err
		}
		if n != 1 {
			return xerr.BadRequest("common.invalidParam")
		}
		return nil
	})
}

func (s *Service) SetChatDisabled(ctx context.Context, id int64, disabled bool) error {
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error { return tx.Member.UpdateOneID(id).SetChatDisabled(disabled).Exec(ctx) })
}
func (s *Service) SetBettingDisabled(ctx context.Context, id int64, disabled bool) error {
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error { return tx.Member.UpdateOneID(id).SetBettingDisabled(disabled).Exec(ctx) })
}
func (s *Service) SetWithdrawalDisabled(ctx context.Context, id int64, disabled bool) error {
	return s.withMemberUpdateTx(ctx, id, func(tx *ent.Tx) error { return tx.Member.UpdateOneID(id).SetWithdrawalDisabled(disabled).Exec(ctx) })
}
