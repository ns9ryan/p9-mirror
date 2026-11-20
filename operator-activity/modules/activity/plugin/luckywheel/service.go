package luckywheel

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"oa.98ent.com/p9/operator-activity/modules/activity/perr"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/audit"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/condition"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/metric"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/payout"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// CrmLookup CRM信息查询
type CrmLookup interface {
	LoadMember(ctx context.Context, memberID int64) (*types.MemberInfo, error)
	// LoadOpBaseInfo(ctx context.Context) (*types.OpBaseInfo, error)
}

// ChanceStore 抽奖机会存储
type ChanceStore interface {
	ExistsWheelChanceGrant(ctx context.Context, key string) (bool, error)
	GrantWheelChances(ctx context.Context, m *repository.WheelChanceGrantModel) error
	GetWheelChanceBalance(ctx context.Context, activityID, memberID int64) (int, error)
	ConsumeWheelChance(ctx context.Context, activityID, memberID int64) error
	LoadOngoingActivities(ctx context.Context, activityType types.ActivityType) ([]*types.Activity, error)
	IdempKeyExists(ctx context.Context, key string) (bool, error)
	SaveOrders(ctx context.Context, orders []*types.ActivityOrder, activity *types.Activity, user *types.MemberInfo) error
	PublishOrderCreated(ctx context.Context, orders []*types.ActivityOrder) error
	GetActivity(ctx context.Context, activityID int64) (*types.Activity, error)
}

// Locker 用户活动锁（可选）
type Locker interface {
	Lock(ctx context.Context, promoID, userID int64) (unlock func(), err error)
}

// Service 幸运转盘业务
type Service struct {
	crmServ CrmLookup
	store   ChanceStore
	deps    *plugin.StrategyDeps
	locker  Locker
	builder *pipeline.Builder
}

func NewService(crmServ CrmLookup, store ChanceStore, deps *plugin.StrategyDeps, locker Locker) *Service {
	return &Service{
		crmServ: crmServ,
		store:   store,
		deps:    deps,
		locker:  locker,
		builder: pipeline.NewBuilder(),
	}
}

// Process 事件预处理
func (s *Service) Process(ctx context.Context, evt *Event) (*Event, *types.MemberInfo, error) {
	if evt == nil {
		return nil, nil, nil
	}
	switch evt.EventType {
	case EventBet, EventDeposit:
	default:
		return nil, nil, nil
	}
	if evt.MemberID <= 0 || evt.Amount <= 0 {
		return nil, nil, perr.Permanentf("luckywheel: invalid event")
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

// BuildPipeline 供抽奖发奖时复用 Audit/Payout；达标送机会不走 Engine 发奖
func (s *Service) BuildPipeline(activity *types.Activity) plugin.ActivityPipeline {
	return &plugin.StaticPipeline{
		// 送机会不校验 ClaimLimit（限制的是奖励订单，在 Spin 时校验）
		CondChain:   buildGrantChain(activity, s.deps),
		RewardCalc:  &noopReward{},
		AuditStrat:  audit.BuildFromSettings(activity),
		PayoutStrat: payout.BuildFromSettings(activity, s.deps.Store, s.deps.Store, s.deps.Workflows),
	}
}

// buildGrantChain 送抽奖机会用的条件链（不含 ClaimLimit）
func buildGrantChain(activity *types.Activity, deps *plugin.StrategyDeps) *condition.Chain {
	return condition.NewChain(
		&condition.ActivityStatusCondition{},
		&condition.ActivityTimeCondition{},
		&condition.FixedLevelCondition{Allowed: activity.AllowedFixedLevel},
		&condition.AutoLevelCondition{Allowed: activity.AllowedAutoLevel},
		&condition.ClientTypeCondition{Allowed: activity.AllowedClients},
		&condition.CycleModeCondition{Manager: deps.CycleMgr},
		&condition.ApplyModeCondition{Mode: activity.ApplyMode, Repo: deps.Store},
		&condition.ApplyLimitCondition{Limits: activity.ApplyLimit, Repo: deps.Store},
		&condition.WheelMetricCondition{Calc: &metric.WheelCalculator{
			Store:        deps.Store,
			Cycle:        deps.CycleMgr,
			Counter:      deps.DepositCounter,
			DepositEvent: EventDeposit,
		}},
	)
}

type noopReward struct{}

func (noopReward) Calculate(*pipeline.EvalContext) ([]types.RewardItem, error) {
	return nil, nil
}

// GrantChances 投注/充值达标后发放抽奖机会
func (s *Service) GrantChances(ctx context.Context, evt *Event, user *types.MemberInfo) error {
	activities, err := s.store.LoadOngoingActivities(ctx, types.ActivityTypeLuckyWheel)
	if err != nil {
		return perr.Retryable(fmt.Errorf("luckywheel: load activities: %w", err))
	}
	for _, activity := range activities {
		if !matchTrigger(activity, evt) {
			continue
		}
		if err := s.grantOne(ctx, activity, evt, user); err != nil {
			return err
		}
	}
	return nil
}

// matchTrigger 按配置匹配事件：可仅投注、仅充值，或两者皆可（满足其一即可）
func matchTrigger(activity *types.Activity, evt *Event) bool {
	// 如果活动配置为空，则不匹配
	if activity.Settings.LuckWheel == nil {
		return false
	}
	// 触发模式
	triggerMode := activity.Settings.LuckWheel.WheelTriggerMode
	switch evt.EventType {
	case EventBet:
		return triggerMode.Has(types.WheelTriggerBet)
	case EventDeposit:
		return triggerMode.Has(types.WheelTriggerDeposit)
	default:
		return false
	}
}

// getSchemas 根据事件类型选择对应的奖励 schema
func getSchemas(activity *types.Activity, evt *Event) []types.RewardSchema {
	var schemas []types.RewardSchema
	switch evt.EventType {
	case EventBet:
		schemas = activity.Settings.LuckWheel.BetSchemas
	case EventDeposit:
		schemas = activity.Settings.LuckWheel.DepositSchemas
	}
	return schemas
}

func (s *Service) grantOne(ctx context.Context, activity *types.Activity, evt *Event, memberInfo *types.MemberInfo) error {
	if s.locker != nil {
		unlock, err := s.locker.Lock(ctx, activity.ID, memberInfo.ID)
		if err != nil {
			return err
		}
		defer unlock()
	}

	pipe := s.BuildPipeline(activity)
	stepLog := pipeline.NewStepLogger(activity.ID, memberInfo.ID)
	evalCtx := &pipeline.EvalContext{
		Context:  ctx,
		Activity: activity,
		Event:    evt,
		Member:   memberInfo,
		StepLog:  stepLog,
	}
	defer stepLog.Flush()

	switch pipe.Conditions().Run(evalCtx) {
	case pipeline.ResultError:
		return perr.Retryablef("luckywheel condition error activity=%d", activity.ID)
	case pipeline.ResultPass:
	default:
		stepLog.Step("condition_chain", "skip", "not passed")
		return nil
	}

	round := evalCtx.Round
	if round <= 0 {
		round, _, _ = s.deps.CycleMgr.CurrentRound(activity, evt.EventTime)
		evalCtx.Round = round
	}

	for _, schema := range getSchemas(activity, evt) {
		if !pipeline.MatchThreshold(evalCtx, schema) || pipeline.FilterThreshold(evalCtx, schema) {
			stepLog.Stepf("grant_chance", "skip", "schema=%d, event_type=%d not match threshold", schema.ID, evt.EventType)
			continue
		}
		chances := int(schema.Amount)
		if chances <= 0 {
			chances = 1
		}
		key := fmt.Sprintf("wheel:chance:%d:%d:%d:%d_%d", activity.ID, memberInfo.ID, round, evt.EventType, schema.ID)
		exists, err := s.store.ExistsWheelChanceGrant(ctx, key)
		if err != nil {
			return perr.Retryable(err)
		}
		if exists {
			continue
		}
		if err := s.store.GrantWheelChances(ctx, &repository.WheelChanceGrantModel{
			OperatorCode: activity.OperatorCode,
			ActivityID:   activity.ID,
			MemberID:     memberInfo.ID,
			Round:        round,
			SchemaID:     schema.ID,
			EventType:    evt.EventType,
			OrderID:      evt.OrderID,
			Chances:      chances,
			CreatedAt:    time.Now(),
		}); err != nil {
			return perr.Classify(err)
		}
		stepLog.Stepf("grant_chance", "pass", "schema=%d chances=%d key=%s", schema.ID, chances, key)
	}
	return nil
}

// SpinResult 抽奖结果
type SpinResult struct {
	Order  *types.ActivityOrder
	Prize  types.RewardSchema
	Remain int
}

// Spin 消耗一次机会并按奖池权重抽奖，创建奖励订单（复用 Audit/Payout）
func (s *Service) Spin(ctx context.Context, activityID int64, memberID int64) (*SpinResult, error) {
	activity, err := s.store.GetActivity(ctx, activityID)
	if err != nil {
		return nil, perr.Classify(fmt.Errorf("luckywheel: load promo: %w", err))
	}
	if activity.Type != types.ActivityTypeLuckyWheel || activity.Status != types.ActivityStatusOngoing {
		return nil, perr.Permanentf("luckywheel: activity %d not available", activityID)
	}
	memberInfo, err := s.crmServ.LoadMember(ctx, memberID)
	if err != nil {
		return nil, perr.Classify(err)
	}

	if s.locker != nil {
		unlock, err := s.locker.Lock(ctx, activityID, memberID)
		if err != nil {
			return nil, err
		}
		defer unlock()
	}

	now := time.Now()
	round, window, _ := s.deps.CycleMgr.CurrentRound(activity, now)

	// 领取/发奖次数限制（与其它活动一致）
	claimCond := &condition.ClaimLimitCondition{Limits: activity.ClaimLimit, Repo: s.deps.Store}
	evalCtx := &pipeline.EvalContext{
		Context:  ctx,
		Activity: activity,
		Member:   memberInfo,
		Round:    round,
		Window:   window,
		Event:    &Event{EventTime: now, MemberID: memberID},
		StepLog:  pipeline.NewStepLogger(activity.ID, memberID),
	}
	if claimCond.Evaluate(evalCtx) != pipeline.ResultPass {
		return nil, perr.Permanentf("luckywheel: claim limit exceeded")
	}

	if err := s.store.ConsumeWheelChance(ctx, activity.ID, memberID); err != nil {
		return nil, perr.Permanentf("luckywheel: %v", err)
	}

	prize, err := drawPrize(activity.Settings.RewardSchemas)
	if err != nil {
		return nil, perr.Permanent(err)
	}

	evt := &Event{
		EventType:    EventSpin,
		OperatorCode: activity.OperatorCode,
		MemberID:     memberID,
		Amount:       prize.Amount,
		EventTime:    now,
	}
	item := types.RewardItem{
		SchemaID:       prize.ID,
		Amount:         prize.Amount,
		Remark:         fmt.Sprintf("luckywheel spin prize schema=%d", prize.ID),
		IdempotencyKey: fmt.Sprintf("wheel:spin:%d:%d:%d:%d:%d", activity.ID, memberID, round, prize.ID, now.UnixNano()),
	}

	pipe := s.BuildPipeline(activity)
	orders := s.builder.BuildFrom(activity, evt, memberInfo, round, []types.RewardItem{item})
	if len(orders) == 0 {
		return nil, perr.Permanentf("luckywheel: build order failed")
	}
	o := orders[0]
	switch pipe.Audit().Decide(o) {
	case audit.DecisionAutoPass:
		o.Status = types.OrderStatusApproved
		o.Remark = "luckywheel auto audit"
		pipe.Payout().Apply(ctx, o)
	default:
		o.Status = types.OrderStatusPendingAudit
		o.Remark = "luckywheel pending audit"
	}

	if err := s.store.SaveOrders(ctx, orders, activity, memberInfo); err != nil {
		return nil, perr.Classify(err)
	}
	if o.Status != types.OrderStatusPendingAudit {
		if err := pipe.Payout().Schedule(ctx, o); err != nil {
			return nil, perr.Retryable(err)
		}
	}
	if err := s.store.PublishOrderCreated(ctx, orders); err != nil {
		logx.Errorf("[luckywheel] publish order created err=%v", err)
	}

	remain, _ := s.store.GetWheelChanceBalance(ctx, activityID, memberID)
	return &SpinResult{Order: o, Prize: prize, Remain: remain}, nil
}

func drawPrize(schemas []types.RewardSchema) (types.RewardSchema, error) {
	if len(schemas) == 0 {
		return types.RewardSchema{}, fmt.Errorf("empty prize pool")
	}
	total := 0
	for _, s := range schemas {
		w := s.Rate
		if w <= 0 {
			w = 1
		}
		total += w
	}
	r := rand.Intn(total)
	cum := 0
	for _, s := range schemas {
		w := s.Rate
		if w <= 0 {
			w = 1
		}
		cum += w
		if r < cum {
			return s, nil
		}
	}
	return schemas[len(schemas)-1], nil
}
