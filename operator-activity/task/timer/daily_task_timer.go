package timer

import (
	"context"
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
	"oa.98ent.com/p9/operator-game/rpc/client/gameorderservice"
)

// DailyTaskTimer 每日下注任务定时触发器
type DailyTaskTimer struct {
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

// NewDailyTaskTimer 创建新的定时器
func NewDailyTaskTimer(
	ctx context.Context,
	db *ent.Client,
	redisQueue *taskqueue.RedisQueue,
	kafkaProducer *kafka.Producer,
	gameRpcClient *game_grpc_client.GameClientManager,
) *DailyTaskTimer {
	timerCtx, cancel := context.WithCancel(ctx)
	return &DailyTaskTimer{
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
func (t *DailyTaskTimer) Start() error {
	// 为每个时区创建定时 goroutine
	for _, task := range t.db.TaskConfig.Query().Where(taskconfig.StatusEQ(constants.TaskConfigStatusEnabled)).AllX(t.ctx) {
		if task.TaskCode != constants.TaskCodeBetAnyGame {
			continue
		}
		t.wg.Add(1)
		go t.runForTimezone(t.ctx, t.redisQueue, task.Timezone, task.OperatorCode, task.TaskCode)
	}

	return nil
}

// Stop 停止定时器
func (t *DailyTaskTimer) Stop() {
	t.cancel()
	t.wg.Wait()
	t.Infof("定时器已停止")
}

// runForTimezone 为特定时区执行定时任务
func (t *DailyTaskTimer) runForTimezone(ctx context.Context, redisQueue *taskqueue.RedisQueue, tz string, operatorCode, taskCode string) {
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
			t.Infof("时区 %s 的调度器已停止", tz)
			return

		default:
			// 计算下一次触发时间（明天 00:00:00）
			now := time.Now().In(loc)
			nextTrigger := time.Date(
				now.Year(),
				now.Month(),
				now.Day()+1,
				0, 0, 0, 0,
				loc,
			)

			waitDuration := nextTrigger.Sub(now)
			t.Infof("时区=%s，下次触发=%s，等待时间=%v", tz, nextTrigger, waitDuration)

			// 等待到触发时间
			select {
			case <-time.After(waitDuration):
				// 触发时间到，执行任务
				t.triggerTasks(ctx, redisQueue, operatorCode, taskCode)

			case <-ctx.Done():
				t.Infof("时区 %s 的调度器已停止", tz)
				return
			}
		}
	}
}

// triggerTasks 触发该时区的所有任务
func (t *DailyTaskTimer) triggerTasks(ctx context.Context, redisQueue *taskqueue.RedisQueue, operatorCode, taskCode string) {
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
	// 获取前一天的时间范围（该时区的 00:00 ~ 23:59:59）
	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Errorf("加载时区 %s 失败，err=%v", tz, err)
		return
	}

	now := time.Now().In(loc)
	yesterday := now.AddDate(0, 0, -1)

	startTime := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		0, 0, 0, 0,
		loc,
	)
	endTime := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		23, 59, 59, 999999999,
		loc,
	)

	// 转换为毫秒时间戳
	startTimeMs := startTime.UnixMilli()
	endTimeMs := endTime.UnixMilli()

	t.Infof("查询下注会员，时区=%s，日期范围=[%s, %s]",
		tz, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 为每个分站调用 RPC 查询会员
	t.processBetTaskForOperator(ctx, redisQueue, task, startTimeMs, endTimeMs)
}

// processBetTaskForOperator 处理单个分站的下注任务
func (t *DailyTaskTimer) processBetTaskForOperator(
	ctx context.Context,
	redisQueue *taskqueue.RedisQueue,
	taskConfig *ent.TaskConfig,
	startTimeMs, endTimeMs int64,
) {
	opCode := taskConfig.OperatorCode
	taskCode := taskConfig.TaskCode

	// 调用游戏 RPC 查询会员
	resp, err := t.gameRpcClient.GetGameOrderServiceClient().QueryBetMembers(ctx, &gameorderservice.QueryBetMembersRequest{
		OperatorCode: opCode,
		StartTime:    startTimeMs,
		EndTime:      endTimeMs,
	})
	if err != nil {
		t.Errorf("查询下注会员失败，分站=%s，err=%v", opCode, err)
		return
	}
	memberIDs := resp.GetMemberIds()
	occurredAt := time.UnixMilli(endTimeMs)
	periodKey := utils.CalculatePeriodKey(taskConfig, occurredAt)

	t.Infof("已查询下注会员，分站=%s，任务码=%s, 周期键=%s，会员数=%d",
		opCode, taskCode, periodKey, len(memberIDs))

	// 为每个会员生产事件
	for _, memberID := range memberIDs {
		event := types.Event{
			KafkaID:      "TEST",
			EventType:    constants.EventTypeValidBetSettled, // 下注完成事件
			OperatorCode: opCode,
			MemberID:     memberID,
			OccurredAt:   occurredAt,
			PublishedAt:  time.Now(),
			RetryCount:   0,
			MaxRetry:     3,
		}

		if err != nil {
			t.Errorf("序列化事件失败，分站=%s，会员ID=%d，err=%v",
				opCode, memberID, err)
			continue
		}
		if err := redisQueue.DispatchTaskEventToBatch(ctx, &event); err != nil {
			t.Errorf("❌ 推送事件到队列失败: 任务码=%s, kafka_id=%s, 分站=%s, err=%v",
				taskCode, event.KafkaID, event.OperatorCode, err)
			return
		}

		// TODO 消息通知Topic
		topic := "msg." + opCode + ".bet.completed"
		if err := t.kafkaProducer.Publish(ctx, topic, event); err != nil {
			t.Errorf("发布事件失败，分站=%s，会员ID=%d，err=%v",
				opCode, memberID, err)
			continue
		}

		t.Debugf("已发布事件，分站=%s，会员ID=%d，任务码=%s",
			opCode, memberID, taskCode)
	}
}
