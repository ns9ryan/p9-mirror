package member

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/agentuser"
	entmember "oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberauth"
	"oa.98ent.com/p9/operator-crm/ent/memberprofile"
)

// GetDetail 查询会员详情及关联信息
func (s *Service) GetDetail(ctx context.Context, memberID int64) (*ent.Member, error) {
	query := s.db.Member.Query().
		Where(entmember.IDEQ(memberID))

	// ==================== 关联信息 ====================

	// 加载所属分层代理, 用于显示代理账号和线路
	query.WithAgent(func(q *ent.AgentUserQuery) {
		q.Select(
			agentuser.FieldID,            // 代理ID
			agentuser.FieldAccount,       // 登录账号, 保留原始大小写
			agentuser.FieldAgentLineCode, // 代理子线路编码
		)
	})

	// 加载会员认证信息, 仅读取登录账号
	query.WithAuth(func(q *ent.MemberAuthQuery) {
		q.Select(
			memberauth.FieldID,       // 认证记录ID
			memberauth.FieldMemberID, // 所属会员ID
			memberauth.FieldUsername, // 会员登录账号
		)
	})

	// 加载会员个人资料, 不读取敏感资料密文
	query.WithProfile(func(q *ent.MemberProfileQuery) {
		q.Select(
			memberprofile.FieldID,             // 个人资料ID
			memberprofile.FieldMemberID,       // 所属会员ID
			memberprofile.FieldAvatarURL,      // 会员头像URL, 使用系统预设资源
			memberprofile.FieldAvatarFrameURL, // 会员头像框URL, 使用系统预设资源
			memberprofile.FieldNickname,       // 会员昵称
			memberprofile.FieldBirthday,       // 会员生日
			memberprofile.FieldGender,         // 性别: 0未设置, 1男, 2女, 3保密
			memberprofile.FieldRegionCode,     // 会员国家或地区编码
			memberprofile.FieldRemark,         // 后台会员备注
		)
	})

	// ==================== 等级与统计 ====================

	// 加载会员自动层级和VIP等级
	query.WithAutoLevel()
	query.WithVipLevel()

	// 加载会员固定层级及分配信息
	query.WithFixedLevelAssignments(func(q *ent.MemberFixedLevelAssignmentQuery) {
		q.WithFixedLevel()
	})

	// 加载会员累计及当前周期统计, 允许统计记录不存在
	query.WithStats()

	m, err := query.Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询会员详情失败: %w", err)
	}

	return m, nil
}
