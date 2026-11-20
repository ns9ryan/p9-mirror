package member

import (
	"context"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/common/xerr"
	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
)

const (
	ContextKindBasic        = "basic"
	ContextKindVerification = "verification"
)

// GetMemberWithContextData 查询会员注册来源及指定关联，租户隔离由 Ent 拦截器处理。
func (s *Service) GetMemberWithContextData(ctx context.Context, id int64, kind string) (*ent.Member, error) {
	if ctxdata.OperatorCodeFromCtx(ctx) == "" {
		return nil, xerr.Unauthorized("common.unauthorized")
	}
	if id <= 0 {
		return nil, xerr.BadRequest("common.invalidParam")
	}
	q := s.db.Member.Query().Where(member.IDEQ(id)).Select(member.FieldID, member.FieldOperatorCode, member.FieldAutoLevelID, member.FieldAgentID, member.FieldAgentAccountID, member.FieldInviterMemberID, member.FieldRegisterIP, member.FieldRegisterDeviceCode)
	// 基本信息加载固定层级分配，验证状态加载所需认证字段。
	switch kind {
	case ContextKindBasic:
		q.WithFixedLevelAssignments(func(q *ent.MemberFixedLevelAssignmentQuery) {
			q.Select(memberfixedlevelassignment.FieldMemberID, memberfixedlevelassignment.FieldFixedLevelID)
		})
	case ContextKindVerification:
		q.WithAuth(func(q *ent.MemberAuthQuery) {
			q.Select(memberauth.FieldID, memberauth.FieldMemberID, memberauth.FieldEmailCiphertext, memberauth.FieldPhoneCiphertext, memberauth.FieldWithdrawalPasswordHash)
		})
	default:
		return nil, xerr.BadRequest("common.invalidParam")
	}
	result, err := q.Only(ctx)
	if ent.IsNotFound(err) {
		return nil, xerr.BadRequest("common.invalidParam")
	}
	return result, err
}
