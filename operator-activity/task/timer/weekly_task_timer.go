package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/taskconfig"
	"oa.98ent.com/p9/operator-activity/task/constants"
	"oa.98ent.com/p9/operator-activity/task/kafka"
	taskqueue "oa.98ent.com/p9/operator-activity/task/queue"
	"oa.98ent.com/p9/operator-activity/task/types"
	"oa.98ent.com/p9/operator-activity/task/utils"
	game_grpc_client "oa.98ent.com/p9/operator-game/pkg/grpc_client"
)

// WeeklyTaskTimer 每周任务定时触发器
type WeeklyTaskTimer struct {
	ctx           context.Context
	cancel        context.CancelFunc
	db            *ent.Client
	redisQueue    *taskqueue.RedisQueue
	kafkaProducer *kafka.Producer
	gameRpcClient *game_grpc_client.GameClientManager

	logx.Logger

	mu sync.RWMutex

	// goroutine 管理
	wg sync.WaitGroup
}

// NewWeeklyTaskTimer 创建新的定时器
func NewWeeklyTaskTimer(
	ctx context.Context,
	db *ent.Client,
	redisQueue *taskqueue.RedisQueue,
	kafkaProducer *kafka.Producer,
	gameRpcClient *game_grpc_client.GameClientManager,
) *WeeklyTaskTimer {
	timerCtx, cancel := context.WithCancel(ctx)
	return &WeeklyTaskTimer{
		ctx:           timerCtx,
		cancel:        cancel,
		db:            db,
		redisQueue:    redisQueue,
		kafkaProducer: kafkaProducer,
		gameRpcClient: gameRpcClient,
		Logger:        logx.WithContext(timerCtx),
	}
}

// Start 启动定时器
func (t *WeeklyTaskTimer) Start() error {
	// 为每个分站的每个任务创建定时 goroutine
	for _, task := range t.db.TaskConfig.Query().Where(taskconfig.StatusEQ(constants.TaskConfigStatusEnabled)).AllX(t.ctx) {
		// PROMO_REGISTER 任务
		if task.TaskCode == constants.TaskCodePromoRegister {
			t.wg.Add(1)
			go t.runForTimezone(t.ctx, t.redisQueue, task.Timezone, task.OperatorCode, task.TaskCode)
		}
	}

	return nil
}

// Stop 停止定时器
func (t *WeeklyTaskTimer) Stop() {
	t.cancel()
	t.wg.Wait()
	t.Infof("定时器已停止")
}

// runForTimezone 为特定时区执行定时任务
func (t *WeeklyTaskTimer) runForTimezone(ctx context.Context, redisQueue *taskqueue.RedisQueue, tz, operatorCode, taskCode string) {
	defer t.wg.Done()

	t.Infof("为时区 %s 启动调度器，任务=%s", tz, taskCode)

	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("无效的时区 %s，err=%v", tz, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			t.Infof("时区 %s 的调度器已停止，任务=%s", tz, taskCode)
			return

		default:
			// 计算下一次触发时间（下周一 00:00:00）
			now := time.Now().In(loc)
			daysUntilNextMonday := (8 - int(now.Weekday())) % 7
			if daysUntilNextMonday == 0 {
				daysUntilNextMonday = 7
			}
			nextTrigger := time.Date(
				now.Year(),
				now.Month(),
				now.Day()+daysUntilNextMonday,
				0, 0, 0, 0,
				loc,
			)

			waitDuration := nextTrigger.Sub(now)
			t.Infof("时区=%s，任务=%s，下次触发=%s，等待时间=%v", tz, taskCode, nextTrigger, waitDuration)

			// 等待到触发时间
			select {
			case <-time.After(waitDuration):
				// 触发时间到，执行任务
				t.triggerTasks(ctx, redisQueue, operatorCode, taskCode)

			case <-ctx.Done():
				t.Infof("时区 %s 的调度器已停止，任务=%s", tz, taskCode)
				return
			}
		}
	}
}

// triggerTasks 触发该时区的任务
func (t *WeeklyTaskTimer) triggerTasks(ctx context.Context, redisQueue *taskqueue.RedisQueue, operatorCode, taskCode string) {
	// 这里查询最新任务
	task, err := redisQueue.GetTaskConfig(ctx, operatorCode, taskCode)
	if err != nil {
		t.Errorf("查询最新任务配置失败，分站=%s，任务码=%s，err=%v", operatorCode, taskCode, err)
		return
	}
	if task.Status != constants.TaskConfigStatusEnabled {
		t.Infof("任务未启用，分站=%s，任务码=%s", operatorCode, taskCode)
		return
	}
	tz := task.Timezone
	t.Infof("为时区 %s 触发任务，任务=%s，分站=%s", tz, taskCode, operatorCode)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("加载时区 %s 失败，err=%v", tz, err)
		return
	}
	// 计算前一周的时间范围
	// 当前时间在本地时区
	now := time.Now().In(loc)
	// 周一是 time.Monday = 1
	// 今天距离这周一有多少天
	daysFromMonday := int(now.Weekday()) - 1
	if daysFromMonday < 0 {
		daysFromMonday = 6 // 周日时，距离这周一是6天
	}

	// 这周一 00:00:00
	thisMonday := time.Date(
		now.Year(),
		now.Month(),
		now.Day()-daysFromMonday,
		0, 0, 0, 0,
		loc,
	)

	// 上周一 00:00:00
	lastMonday := thisMonday.AddDate(0, 0, -7)
	// 上周日 23:59:59
	lastSunday := thisMonday.AddDate(0, 0, -1)
	lastSundayEnd := time.Date(
		lastSunday.Year(),
		lastSunday.Month(),
		lastSunday.Day(),
		23, 59, 59, 999999999,
		loc,
	)

	startTimeMs := lastMonday.UnixMilli()
	endTimeMs := lastSundayEnd.UnixMilli()

	t.Infof("计算统计周期，分站=%s，任务=%s，周期范围=[%s, %s]",
		operatorCode, taskCode, lastMonday.Format("2006-01-02 15:04:05"), lastSundayEnd.Format("2006-01-02 15:04:05"))

	// 根据任务类型调用不同的处理器
	switch taskCode {
	case constants.TaskCodePromoRegister:
		t.processPromoRegisterTaskForOperator(ctx, redisQueue, task, startTimeMs, endTimeMs)
	default:
		t.Errorf("未知的周期任务类型，分站=%s，任务=%s", operatorCode, taskCode)
	}
}

// processPromoRegisterTaskForOperator 处理单个分站的邀请任务
func (t *WeeklyTaskTimer) processPromoRegisterTaskForOperator(
	ctx context.Context,
	redisQueue *taskqueue.RedisQueue,
	taskConfig *ent.TaskConfig,
	startTimeMs, endTimeMs int64,
) {
	opCode := taskConfig.OperatorCode
	taskCode := taskConfig.TaskCode
	completionAmount := taskConfig.CompletionAmount

	t.Infof("开始查询邀请符合条件的会员，分站=%s，完成条件=%.2f %s，时间范围=[%d, %d]",
		opCode, float64(completionAmount)/100, startTimeMs, endTimeMs)

	// // 调用 RPC 查询邀请符合条件的会员（邀请人ID）
	// // gRPC 返回的是邀请人ID，即获得奖励的人
	// resp, err := t.gameRpcClient.GetInvServiceClient().QueryMembers(ctx, &gameorderservice.QueryBetMembersRequest{
	// 	OperatorCode: opCode,
	// 	Amount:       completionAmount,
	// 	StartTime:    startTimeMs,
	// 	EndTime:      endTimeMs,
	// })
	// if err != nil {
	// 	t.Errorf("❌ 查询邀请会员失败，分站=%s，err=%v", opCode, err)
	// 	return
	// }

	// memberIDs := resp.GetMemberIds()
	memberIDs := []int64{59, 60, 61} // 临时使用会员ID列表，实际应从RPC获取
	occurredAt := time.UnixMilli(endTimeMs)
	periodKey := utils.CalculatePeriodKey(taskConfig, occurredAt)

	t.Infof("✅ 已查询邀请会员，分站=%s，任务码=%s，周期键=%s，会员数=%d",
		opCode, taskCode, periodKey, len(memberIDs))

	if len(memberIDs) == 0 {
		t.Infof("⚠️ 本周期没有符合条件的邀请会员，分站=%s", opCode)
		return
	}

	// 为每个符合条件的邀请人生产事件
	for _, inviterID := range memberIDs {
		event := types.Event{
			KafkaID:      fmt.Sprintf("promo_%s_%d_%s", opCode, inviterID, periodKey),
			EventType:    constants.EventTypeReferralRegisteredDeposit, // 邀请会员注册并充值事件
			OperatorCode: opCode,
			MemberID:     inviterID, // 邀请人ID（获得奖励的人）
			OccurredAt:   occurredAt,
			PublishedAt:  time.Now(),
			RetryCount:   0,
			MaxRetry:     3,
		}

		// 推送到 Redis 队列
		if err := redisQueue.DispatchTaskEventToBatch(ctx, &event); err != nil {
			t.Errorf("❌ 推送事件到队列失败: task_code=%s, opcode=%s, inviter_id=%d, err=%v",
				taskCode, opCode, inviterID, err)
			continue
		}

		// 发布到 Kafka topic
		topic := "msg." + opCode + ".promo.completed"
		if err := t.kafkaProducer.Publish(ctx, topic, event); err != nil {
			t.Errorf("❌ 发布事件到Kafka失败，分站=%s，邀请人ID=%d，Topic=%s，err=%v",
				opCode, inviterID, topic, err)
			continue
		}

		t.Infof("✅ 已发布邀请事件，分站=%s，邀请人ID=%d，任务码=%s，周期=%s",
			opCode, inviterID, taskCode, periodKey)
	}

	t.Infof("✅ 邀请任务处理完成，分站=%s，总会员数=%d", opCode, len(memberIDs))
}
