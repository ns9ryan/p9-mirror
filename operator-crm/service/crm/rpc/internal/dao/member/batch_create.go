package member

import (
	"context"
	"fmt"
	"strings"

	"oa.98ent.com/p9/operator-crm/ent"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
)

const memberInsertBatchSize = 200

// BatchCreateTx 批量写入会员和初始化记录，事务由调用方提交。
func (s *Service) BatchCreateTx(ctx context.Context, tx *ent.Tx, params []CreateParams) ([]*ent.Member, error) {
	// 整批共用当前分站的默认自动层级和 VIP0。
	autoLevel, err := s.autoLevel.GetDefaultTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("查询默认自动层级失败: %w", err)
	}
	vipLevel, err := s.vipLevel.GetDefaultTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("查询默认VIP等级失败: %w", err)
	}

	members := make([]*ent.Member, 0, len(params))
	// 每组最多 200 个会员，控制单条 SQL 的参数数量。
	for start := 0; start < len(params); start += memberInsertBatchSize {
		batch := params[start:min(start+memberInsertBatchSize, len(params))]
		creates := make([]*ent.MemberCreate, len(batch))
		for i, param := range batch {
			creates[i] = buildMember(tx, param, autoLevel.ID, vipLevel.ID)
		}
		created, err := tx.Member.CreateBulk(creates...).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("批量创建会员主体失败: %w", err)
		}

		// 使用批量插入返回的主键，为每个会员生成认证、资料和 VIP 进度记录。
		auths := make([]*ent.MemberAuthCreate, len(batch))
		profiles := make([]*ent.MemberProfileCreate, len(batch))
		progresses := make([]*ent.MemberVipProgressCreate, len(batch))
		for i, param := range batch {
			id := created[i].ID
			auths[i] = buildMemberAuth(tx, param, id)
			profiles[i] = buildMemberProfile(tx, param, id)
			progresses[i] = tx.MemberVipProgress.Create().SetMemberID(id)
		}
		if err := tx.MemberAuth.CreateBulk(auths...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("批量创建会员认证信息失败: %w", err)
		}
		if err := tx.MemberProfile.CreateBulk(profiles...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("批量创建会员个人资料失败: %w", err)
		}
		if err := tx.MemberVipProgress.CreateBulk(progresses...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("批量创建会员VIP晋级进度失败: %w", err)
		}
		members = append(members, created...)
	}
	return members, nil
}

// IdentifierConflicts 记录候选会员标识中已被使用的值。
type IdentifierConflicts struct {
	Accounts    map[string]bool
	Codes       map[string]bool
	InviteCodes map[string]bool
}

// FindIdentifierConflictsTx 批量检查当前分站已占用的账号、会员编号和邀请码。
func (*Service) FindIdentifierConflictsTx(ctx context.Context, tx *ent.Tx, accounts, codes, invites []string) (IdentifierConflicts, error) {
	result := IdentifierConflicts{Accounts: map[string]bool{}, Codes: map[string]bool{}, InviteCodes: map[string]bool{}}
	auths, err := tx.MemberAuth.Query().Where(memberauth.UsernameIn(accounts...)).Select(memberauth.FieldUsername).All(ctx)
	if err != nil {
		return result, err
	}
	// 登录账号字段使用 citext，返回的冲突账号统一转为小写供调用方比较。
	for _, auth := range auths {
		if auth.Username != nil {
			result.Accounts[strings.ToLower(*auth.Username)] = true
		}
	}
	members, err := tx.Member.Query().Where(entmember.Or(entmember.CodeIn(codes...), entmember.InviteCodeIn(invites...))).
		Select(entmember.FieldCode, entmember.FieldInviteCode).All(ctx)
	if err != nil {
		return result, err
	}
	for _, member := range members {
		result.Codes[member.Code] = true
		result.InviteCodes[member.InviteCode] = true
	}
	return result, nil
}
