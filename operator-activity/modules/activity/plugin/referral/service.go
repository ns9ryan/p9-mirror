package referral

import (
	"context"
	"fmt"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/audit"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/condition"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/metric"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/payout"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/reward"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
)

// CrmLookup CRM信息查询
type CrmLookup interface {
	LoadMember(ctx context.Context, memberID int64) (*types.MemberInfo, error)
}

// Service 推荐活动业务服务
type Service struct {
	crmServ CrmLookup
	deps    *plugin.StrategyDeps
}

func NewService(crmServ CrmLookup, deps *plugin.StrategyDeps) *Service {
	return &Service{crmServ: crmServ, deps: deps}
}

// Process 推荐事件业务预处理，返回补全后的 Event 与奖励归属用户（邀请人）
func (s *Service) Process(ctx context.Context, evt *Event) (*Event, *types.MemberInfo, error) {
	if evt == nil {
		return nil, nil, nil
	}
	switch evt.EventType {
	case EventRegister:
		return s.processRegister(ctx, evt)
	case EventDeposit:
		return s.processReferralDeposit(ctx, evt)
	case EventBet:
		return s.processReferralBetting(ctx, evt)
	default:
		return evt, nil, nil
	}
}

// BuildPipeline 组装推荐活动评估流水线
func (s *Service) BuildPipeline(promo *types.Activity) plugin.ActivityPipeline {
	settings := promo.Settings
	chain := condition.BuildBaseChain(promo, condition.BaseDeps{
		CycleMgr: s.deps.CycleMgr,
		Apps:     s.deps.Store,
		Limits:   s.deps.Store,
	})
	chain.Add(&condition.ReferralCondition{
		Schemas: settings.RewardSchemas,
		Calc:    &metric.ReferralCalculator{Store: s.deps.Store, Cycle: s.deps.CycleMgr},
	})

	return &plugin.StaticPipeline{
		CondChain:   chain,
		RewardCalc:  reward.BuildFromSettings(promo),
		AuditStrat:  audit.BuildFromSettings(promo),
		PayoutStrat: payout.BuildFromSettings(promo, s.deps.Store, s.deps.Store, s.deps.Workflows),
	}
}

func (s *Service) processRegister(ctx context.Context, evt *Event) (*Event, *types.MemberInfo, error) {
	if evt.InviterID <= 0 {
		return nil, nil, nil // 无邀请人，跳过
	}
	inviter, err := s.crmServ.LoadMember(ctx, evt.InviterID)
	if err != nil {
		return nil, nil, perr.Classify(fmt.Errorf("referral: load inviter %d: %w", evt.InviterID, err))
	}
	// 奖励归属邀请人
	evt.InviterID = inviter.ID
	evt.OperatorCode = inviter.OperatorCode
	return evt, inviter, nil
}

func (s *Service) processReferralDeposit(ctx context.Context, evt *Event) (*Event, *types.MemberInfo, error) {
	// 被推荐人充值达标时，查邀请人并触发邀请人的推荐奖励评估
	depositUserID := evt.MemberID
	if depositUserID <= 0 {
		return nil, nil, perr.Permanentf("referral deposit: missing member_id")
	}
	user, err := s.crmServ.LoadMember(ctx, depositUserID)
	if err != nil {
		return nil, nil, perr.Classify(fmt.Errorf("referral: load deposit member %d: %w", depositUserID, err))
	}
	if user.InviterID <= 0 {
		return nil, nil, nil
	}
	inviter, err := s.crmServ.LoadMember(ctx, user.InviterID)
	if err != nil {
		return nil, nil, perr.Classify(fmt.Errorf("referral: load inviter %d: %w", user.InviterID, err))
	}
	evt.InviterID = inviter.ID
	evt.OperatorCode = inviter.OperatorCode
	return evt, inviter, nil
}

func (s *Service) processReferralBetting(ctx context.Context, evt *Event) (*Event, *types.MemberInfo, error) {
	// 被推荐人投注达标时，查邀请人并触发邀请人的推荐奖励评估
	bettingUserID := evt.MemberID
	if bettingUserID <= 0 {
		return nil, nil, perr.Permanentf("referral betting: missing member_id")
	}
	memberInfo, err := s.crmServ.LoadMember(ctx, bettingUserID)
	if err != nil {
		return nil, nil, perr.Classify(fmt.Errorf("referral: load betting user %d: %w", bettingUserID, err))
	}
	if memberInfo.InviterID <= 0 {
		return nil, nil, nil
	}
	inviterInfo, err := s.crmServ.LoadMember(ctx, memberInfo.InviterID)
	if err != nil {
		return nil, nil, perr.Classify(fmt.Errorf("referral: load inviter %d: %w", memberInfo.InviterID, err))
	}
	evt.InviterID = inviterInfo.ID
	evt.OperatorCode = inviterInfo.OperatorCode
	return evt, inviterInfo, nil
}
