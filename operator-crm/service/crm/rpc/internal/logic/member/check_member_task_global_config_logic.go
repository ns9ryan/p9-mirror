package memberlogic

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-crm/ent"
	"oa.98ent.com/p9/operator-crm/ent/member"
	"oa.98ent.com/p9/operator-crm/ent/memberfixedlevelassignment"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/internal/svc"
	"oa.98ent.com/p9/operator-crm/service/crm/rpc/pb/crmrpc/memberpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckMemberTaskGlobalConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckMemberTaskGlobalConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckMemberTaskGlobalConfigLogic {
	return &CheckMemberTaskGlobalConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查会员是否符合当前全局任务配置
func (l *CheckMemberTaskGlobalConfigLogic) CheckMemberTaskGlobalConfig(in *memberpb.CheckMemberTaskGlobalConfigReq) (*memberpb.CheckMemberTaskGlobalConfigResp, error) {
	l.Infof("开始检查会员全局任务配置: memberId=%d", in.MemberId)

	// 获取会员信息，包含自动层级和固定层级分配关系
	memberInfo, err := l.getMemberWithLevels(l.ctx, in.MemberId)
	if err != nil {
		l.Errorf("获取会员信息失败: memberId=%d, err=%v", in.MemberId, err)
		return &memberpb.CheckMemberTaskGlobalConfigResp{
			IsEligible: false,
			Reason:     "获取会员信息失败",
		}, nil
	}

	if memberInfo == nil {
		l.Errorf("会员不存在: memberId=%d", in.MemberId)
		return &memberpb.CheckMemberTaskGlobalConfigResp{
			IsEligible: false,
			Reason:     "会员不存在",
		}, nil
	}
	l.Infof("成功获取会员信息: memberId=%d, autoLevelId=%d", in.MemberId, memberInfo.AutoLevelID)

	// 检查层级符合性
	// 一个用户只会有一种层级：要么有固定层级，要么有自动层级
	l.Infof("开始检查会员层级符合性: memberId=%d, 固定层级IDs=%v, 自动层级IDs=%v", in.MemberId, in.Config.FixedLevelIds, in.Config.AutoLevelIds)
	isLevelEligible, levelReason := l.checkMemberLevelEligibility(memberInfo, in.Config)
	if !isLevelEligible {
		l.Errorf("会员层级检查不通过: memberId=%d, reason=%s", in.MemberId, levelReason)
		return &memberpb.CheckMemberTaskGlobalConfigResp{
			IsEligible: false,
			Reason:     levelReason,
		}, nil
	}
	l.Infof("会员层级检查通过: memberId=%d", in.MemberId)

	l.Infof("会员全局任务配置检查完成，符合条件: memberId=%d", in.MemberId)
	return &memberpb.CheckMemberTaskGlobalConfigResp{
		IsEligible: true,
	}, nil
}

// getMemberWithLevels 获取会员信息，包含自动层级和固定层级分配
func (l *CheckMemberTaskGlobalConfigLogic) getMemberWithLevels(ctx context.Context, memberID int64) (*ent.Member, error) {
	query := l.svcCtx.DB.Member.Query().
		Where(member.IDEQ(memberID)).
		WithFixedLevelAssignments(func(q *ent.MemberFixedLevelAssignmentQuery) {
			q.Select(memberfixedlevelassignment.FieldFixedLevelID)
		})

	memberInfo, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return memberInfo, nil
}

// checkMemberLevelEligibility 检查会员层级是否符合配置
// 逻辑：
// 1. 如果会员有固定层级分配，检查所有的fixed_level_id是否都在config.fixed_level_ids中
// 2. 如果会员没有固定层级分配，检查auto_level_id是否在config.auto_level_ids中
func (l *CheckMemberTaskGlobalConfigLogic) checkMemberLevelEligibility(memberInfo *ent.Member, config *memberpb.TaskGlobalConfig) (bool, string) {
	// 获取会员的固定层级分配
	fixedLevelAssignments := memberInfo.Edges.FixedLevelAssignments
	fixedLevelIDSet := make(map[int64]bool)
	for _, assignment := range fixedLevelAssignments {
		fixedLevelIDSet[assignment.FixedLevelID] = true
	}

	// 情况1：会员有固定层级分配
	if len(fixedLevelIDSet) > 0 {
		// 检查所有的fixed_level_id是否都在config.fixed_level_ids中
		configFixedLevelIDSet := make(map[int64]bool)
		for _, id := range config.FixedLevelIds {
			configFixedLevelIDSet[id] = true
		}

		for fixedLevelID := range fixedLevelIDSet {
			if !configFixedLevelIDSet[fixedLevelID] {
				return false, fmt.Sprintf("会员的固定层级ID %d 不在配置允许的固定层级范围内", fixedLevelID)
			}
		}
		return true, ""
	}

	// 情况2：会员没有固定层级，检查自动层级
	if memberInfo.AutoLevelID > 0 {
		configAutoLevelIDSet := make(map[int64]bool)
		for _, id := range config.AutoLevelIds {
			configAutoLevelIDSet[id] = true
		}

		if !configAutoLevelIDSet[memberInfo.AutoLevelID] {
			return false, fmt.Sprintf("会员的自动层级ID %d 不在配置允许的自动层级范围内", memberInfo.AutoLevelID)
		}
		return true, ""
	}

	// 情况3：会员既没有固定层级也没有自动层级
	return false, "会员没有指定层级"
}
