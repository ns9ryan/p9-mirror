package plugin

import (
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/audit"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/condition"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/metric"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/payout"
	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/reward"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"
	workerOrder "oa.98ent.com/p9/operator-activity/modules/worker/order"
)

// PipelineFactory 由插件 Service 实现：按活动配置组装本业务流水线
type PipelineFactory interface {
	BuildPipeline(activity *types.Activity) ActivityPipeline
}

// ActivityPipeline 单个活动实例的评估/发奖策略集合。
// 达标指标由各自条件在评估时计算，并写入 EvalContext。
type ActivityPipeline interface {
	Conditions() *condition.Chain
	Rewards() reward.Calculator
	Audit() audit.Strategy
	Payout() payout.Strategy
}

// StrategyStore 组装流水线所需的存储能力
type StrategyStore interface {
	metric.MetricStore
	condition.ApplicationChecker
	condition.LimitChecker
	payout.ScheduleStore
	payout.PayoutPublisher
}

// StrategyDeps 插件组装条件链/指标/派发时的共享依赖
type StrategyDeps struct {
	Store          StrategyStore
	CycleMgr       *pipeline.Manager
	DepositCounter metric.DepositWindowCounter // 可选；nil 时充值指标直查 DB
	Workflows      payout.WorkflowStarter      // 可选；nil 时不启 Temporal
}

func NewStrategyDeps(store StrategyStore, depositCounter metric.DepositWindowCounter) *StrategyDeps {
	return &StrategyDeps{
		Store:          store,
		CycleMgr:       pipeline.NewManager(),
		DepositCounter: depositCounter,
		Workflows:      workerOrder.NewStarter(),
	}
}

// StaticPipeline PromoPipeline 的默认实现
type StaticPipeline struct {
	CondChain   *condition.Chain
	RewardCalc  reward.Calculator
	AuditStrat  audit.Strategy
	PayoutStrat payout.Strategy
}

func (p *StaticPipeline) Conditions() *condition.Chain { return p.CondChain }
func (p *StaticPipeline) Rewards() reward.Calculator   { return p.RewardCalc }
func (p *StaticPipeline) Audit() audit.Strategy        { return p.AuditStrat }
func (p *StaticPipeline) Payout() payout.Strategy      { return p.PayoutStrat }
