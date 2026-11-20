package betting

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
	// LoadOpBaseInfo(ctx context.Context) (*types.OpBaseInfo, error)
}

// Service 投注活动业务服务
type Service struct {
	crmServ CrmLookup
	deps    *plugin.StrategyDeps
}

func NewService(crmServ CrmLookup, deps *plugin.StrategyDeps) *Service {
	return &Service{crmServ: crmServ, deps: deps}
}

// Process 投注结算事件业务预处理：校验 → 补全 → 返回可派发的 Event 与已加载用户
func (s *Service) Process(ctx context.Context, evt *Event) (*Event, *types.MemberInfo, error) {
	if evt == nil {
		return nil, nil, nil
	}
	if err := validateBetEvent(evt); err != nil {
		return nil, nil, err
	}
	// 加载会员信息
	memberInfo, err := s.crmServ.LoadMember(ctx, evt.MemberID)
	if err != nil {
		return nil, nil, perr.Classify(fmt.Errorf("betting: load member %d: %w", evt.MemberID, err))
	}

	// 补全事件信息
	evt.OperatorCode = memberInfo.OperatorCode

	return evt, memberInfo, nil
}

// BuildPipeline 组装投注活动评估流水线
func (s *Service) BuildPipeline(activity *types.Activity) plugin.ActivityPipeline {
	chain := condition.BuildBaseChain(activity, condition.BaseDeps{
		CycleMgr: s.deps.CycleMgr,
		Apps:     s.deps.Store,
		Limits:   s.deps.Store,
	})
	chain.Add(&condition.BetCondition{
		MinAmount:      0,
		AllowedGameIds: activity.AllowedGameIds,
		Calc:           &metric.BettingCalculator{Store: s.deps.Store, Cycle: s.deps.CycleMgr},
	})

	return &plugin.StaticPipeline{
		CondChain:   chain,
		RewardCalc:  reward.BuildFromSettings(activity),
		AuditStrat:  audit.BuildFromSettings(activity),
		PayoutStrat: payout.BuildFromSettings(activity, s.deps.Store, s.deps.Store, s.deps.Workflows),
	}
}

func validateBetEvent(evt *Event) error {
	if evt.MemberID <= 0 {
		return perr.Permanentf("betting: invalid member_id")
	}
	if evt.OrderID <= 0 {
		return perr.Permanentf("betting: invalid order_id")
	}
	if evt.Amount <= 0 {
		return perr.Permanentf("betting: invalid bet_amount")
	}
	if evt.GetEventType() != EventSettled {
		return perr.Permanentf("betting: unsupported event type")
	}
	return nil
}

// MatchGameType 校验游戏类型是否在活动允许列表中（空列表表示不限制）
func MatchGameType(allowed []int, gameType int) bool {
	if len(allowed) == 0 {
		return true
	}
	if gameType == 0 {
		return false
	}
	for _, t := range allowed {
		if t == gameType {
			return true
		}
	}
	return false
}
