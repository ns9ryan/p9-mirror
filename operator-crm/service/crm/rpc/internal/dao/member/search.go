package member

import (
	"context"
	"fmt"
	"strconv"

	"entgo.io/ent/dialect/sql"

	"oa.98ent.com/p9/common/ctxdata"
	"oa.98ent.com/p9/operator-crm/ent"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberprofile"
	"oa.98ent.com/p9/operator-crm/ent/predicate"
)

const maxSearchLimit = 50

// SearchByCode 按关键字模糊搜索会员，匹配编号/账号/昵称。
// 若 keyword 能解析为正整数，则在 OR 条件中额外追加 member.IDEQ 精确匹配，
// 其余三个模糊条件照常保留，二者并集返回。
// 返回 id、编号、账号、昵称，结果按 ID 倒序排列，上限 maxSearchLimit 条。
func (s *Service) SearchByCode(ctx context.Context, keyword string, limit int) ([]*ent.Member, error) {
	if limit <= 0 || limit > maxSearchLimit {
		limit = 20
	}
	operator := ctxdata.OperatorCodeFromCtx(ctx)
	if operator == "" {
		return nil, fmt.Errorf("missing operator scope")
	}

	orPreds := []predicate.Member{
		entmember.CodeContainsFold(keyword),
		entmember.HasAuthWith(
			memberauth.OperatorCodeEQ(operator),
			memberauth.UsernameContainsFold(keyword),
		),
		entmember.HasProfileWith(
			memberprofile.OperatorCodeEQ(operator),
			memberprofile.NicknameContainsFold(keyword),
		),
	}
	// keyword 能解析为正整数时，追加 ID 精确匹配（不替换模糊条件）
	if id, err := strconv.ParseInt(keyword, 10, 64); err == nil && id > 0 {
		orPreds = append(orPreds, entmember.IDEQ(id))
	}

	rows, err := s.db.Member.Query().
		Where(
			entmember.OperatorCodeEQ(operator),
			entmember.Or(orPreds...),
		).
		Select(entmember.FieldID, entmember.FieldCode).
		WithAuth(func(q *ent.MemberAuthQuery) {
			q.Select(memberauth.FieldMemberID, memberauth.FieldUsername)
		}).
		WithProfile(func(q *ent.MemberProfileQuery) {
			q.Select(memberprofile.FieldMemberID, memberprofile.FieldNickname)
		}).
		Order(entmember.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("搜索会员失败: %w", err)
	}
	return rows, nil
}
