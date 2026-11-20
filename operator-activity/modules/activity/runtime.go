package promotion

import (
	"context"

	"oa.98ent.com/p9/operator-activity/internal/svc"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/betting"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository/crm"

	"oa.98ent.com/p9/operator-activity/modules/activity/pipeline/metric"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/deposit"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/luckywheel"
	"oa.98ent.com/p9/operator-activity/modules/activity/plugin/referral"
	"oa.98ent.com/p9/operator-activity/modules/activity/repository"
	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// Runtime 活动系统运行时，聚合引擎、调度器、仓储
type Runtime struct {
	Engine     *Engine
	Repo       *repository.Repository
	Scheduler  *Scheduler
	LuckyWheel *luckywheel.Service // 抽奖 Spin / 查余额
	pipelines  map[types.ActivityType]plugin.PipelineFactory
}

// InitOptions 可选依赖
type InitOptions struct {
	RedisAddr string // 非空且可连通时：启用充值窗口计数器 + 用户活动分布式锁
}

// Init 初始化活动系统，并将插件注册到 Scheduler
func Init(ctx *svc.ServiceContext, brokers ...string) (*Runtime, error) {
	return InitWithOptions(ctx, InitOptions{}, brokers...)
}

// InitWithOptions 带可选 Redis 等依赖的初始化
func InitWithOptions(ctx *svc.ServiceContext, opts InitOptions, brokers ...string) (*Runtime, error) {
	repo := repository.New(ctx.DB, brokers...)
	crmServ := crm.NewCrmServ(ctx)

	var depositCounter metric.DepositWindowCounter
	var locker UserPromoLocker
	if opts.RedisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: opts.RedisAddr})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			logx.Errorf("[activity] redis ping failed addr=%s err=%v, deposit counter/locker disabled", opts.RedisAddr, err)
		} else {
			depositCounter = repository.NewRedisDepositCounter(rdb)
			locker = repository.NewRedisUserPromoLocker(rdb)
			logx.Infof("[activity] redis enabled addr=%s (deposit counter + user-promo lock)", opts.RedisAddr)
		}
	}

	engine := NewEngine(repo, locker)
	sched := NewScheduler(ctx.ServiceName)
	deps := plugin.NewStrategyDeps(repo, depositCounter)

	// 各插件独立 Kafka Group，同 topic 不再保证进程内处理顺序
	depPlugin := deposit.New(engine, crmServ, deps)
	betPlugin := betting.New(engine, crmServ, deps)
	refPlugin := referral.New(engine, crmServ, deps)
	lw := luckywheel.New(crmServ, repo, deps, locker)
	sched.Register(depPlugin)
	sched.Register(betPlugin)
	sched.Register(refPlugin)
	sched.Register(lw)

	pipelines := map[types.ActivityType]plugin.PipelineFactory{
		types.ActivityTypeDeposit:        depPlugin.Service(),
		types.ActivityTypeWagering:       betPlugin.Service(),
		types.ActivityTypeReferralMember: refPlugin.Service(),
		types.ActivityTypeLuckyWheel:     lw.Service(),
	}

	return &Runtime{
		Engine:     engine,
		Repo:       repo,
		Scheduler:  sched,
		LuckyWheel: lw.Service(),
		pipelines:  pipelines,
	}, nil
}

// PipelineFactory 按活动类型取对应插件流水线工厂（审核通过时组装 Payout 等）
func (rt *Runtime) PipelineFactory(activityType types.ActivityType) plugin.PipelineFactory {
	if rt == nil {
		return nil
	}
	return rt.pipelines[activityType]
}
