package svc

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"

	corequeue "github.com/zeromicro/go-zero/core/queue"
	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/ent/migrate"
	_ "oa.98ent.com/p9/operator-activity/ent/runtime"
	"oa.98ent.com/p9/operator-activity/task/config"
	"oa.98ent.com/p9/operator-activity/task/kafka"
	"oa.98ent.com/p9/operator-activity/task/queue"
	"oa.98ent.com/p9/operator-activity/task/timer"

	game_grpc_client "oa.98ent.com/p9/operator-game/pkg/grpc_client"
)

type ServiceContext struct {
	Config     config.Config
	DB         *ent.Client       // Ent数据库客户端
	RedisQueue *queue.RedisQueue // Redis队列
	// Kafka生产者（全局共享）
	KafkaProducer *kafka.Producer // Kafka生产者

	// Kafka消费者（全局唯一）
	KafkaQueueMessages []corequeue.MessageQueue // Kafka消费者（订阅所有分站的topics）

	// Redis客户端（全局共享）
	RedisClient *redis.Client // Redis客户端

	// 每日任务定时器
	DailyTaskTimer  *timer.DailyTaskTimer  // 每日任务定时触发器
	WeeklyTaskTimer *timer.WeeklyTaskTimer // 每周任务定时触发器
}

func NewServiceContext(c config.Config) *ServiceContext {
	driver, err := c.DatabaseConf.NewDriver()
	logx.Must(err)

	entOpts := []ent.Option{
		ent.Log(logx.Info),
		ent.Driver(driver),
	}
	if c.IsDebug() {
		entOpts = append(entOpts, ent.Debug())
	}

	db := ent.NewClient(entOpts...)

	logx.Must(db.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithForeignKeys(false),
	))

	return InitializeTaskServiceContext(c, db)
}

// Close 关闭服务上下文中的所有资源
func (sc *ServiceContext) Close() error {
	var err error

	// 关闭下注任务定时器
	if sc.DailyTaskTimer != nil {
		sc.DailyTaskTimer.Stop()
		logx.Info("DailyTaskTimer closed")
	}

	// 关闭全局 Kafka Consumer（全局唯一）
	if sc.KafkaQueueMessages != nil {
		kafka.StopKq(sc.KafkaQueueMessages)
	}

	if sc.KafkaProducer != nil {
		sc.KafkaProducer.Close()
		logx.Info("Kafka Producer closed")
	}

	time.Sleep(2 * time.Second) // 等待最后一笔kafka消费结束，确保消息处理完成。不怕重复消费，有幂等性保障

	return err
}

func InitializeTaskServiceContext(c config.Config, db *ent.Client) *ServiceContext {
	ctx := context.Background()
	redisCfg := c.RedisConf
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Address,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	// 测试连接
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logx.Errorf("连接Redis失败: %v", err)
	} else {
		logx.Info("连接任务Redis成功")
	}

	// 初始化Kafka生产者（全局共享）
	kafkaProducer := kafka.NewProducer(c.KqBrokers)
	operatorBaseClient := zrpc.MustNewClient(c.OperatorBaseRpc)
	// 初始化所有分站的默认任务配置
	config.InitializeAllDefaultConfigs(ctx, db, operatorBaseClient)
	redisQueue := queue.NewRedisQueue(ctx, redisClient, db, c.Name)
	// 为该 opcode 创建独立的 Queue Consumer
	queueConsumer := queue.NewQueueConsumer(ctx, redisQueue, db, 1*time.Second)

	queueConsumer.Start(ctx)
	// 创建全局唯一的 Kafka Consumer（订阅所有分站的 topics）
	kafkaQueueMessages := kafka.StartKq(
		ctx,
		redisQueue,
		c.KqBrokers,
	)

	var dailyTaskTimer *timer.DailyTaskTimer
	var weeklyTaskTimer *timer.WeeklyTaskTimer
	if c.EnableTaskTimer {
		// 初始化游戏 RPC 客户端
		operatorGameClient, err := game_grpc_client.NewOperatorGameClientManager(c.OperatorGameRpc)
		if err != nil {
			logx.Errorf("初始化游戏 RPC 客户端失败: %v, 定时任务将无法工作", err)
			// 继续运行，但定时器可能无法工作
		}
		// 初始化每日下注任务定时器
		dailyTaskTimer = timer.NewDailyTaskTimer(
			ctx,
			db,
			redisQueue,
			kafkaProducer,
			operatorGameClient,
		)

		// 启动定时器
		if err := dailyTaskTimer.Start(); err != nil {
			logx.Errorf("启动每日任务定时器失败: %v", err)
		}

		weeklyTaskTimer = timer.NewWeeklyTaskTimer(
			ctx,
			db,
			redisQueue,
			kafkaProducer,
			operatorGameClient,
		)

		// 启动每周任务定时器
		if err := weeklyTaskTimer.Start(); err != nil {
			logx.Errorf("启动每周任务定时器失败: %v", err)
		}
	} else {
		logx.Info("定时器未启用")
	}

	// 启动定时打印指标（每 120 秒打印一次）
	queue.PrintMetricsPeriodicly(120 * time.Second)
	return &ServiceContext{
		Config:             c,
		DB:                 db,
		KafkaProducer:      kafkaProducer,
		KafkaQueueMessages: kafkaQueueMessages,
		RedisClient:        redisClient,
		RedisQueue:         redisQueue,
		DailyTaskTimer:     dailyTaskTimer,
		WeeklyTaskTimer:    weeklyTaskTimer,
	}
}

func StopTaskServiceContext(sc *ServiceContext) error {
	var err error

	if sc.DailyTaskTimer != nil {
		sc.DailyTaskTimer.Stop()
		logx.Info("Daily Task Timer stopped")
	}

	if sc.WeeklyTaskTimer != nil {
		sc.WeeklyTaskTimer.Stop()
		logx.Info("Weekly Task Timer stopped")
	}

	if sc.KafkaProducer != nil {
		sc.KafkaProducer.Close()
		logx.Info("Kafka Producer closed")
	}

	if sc.RedisClient != nil {
		err = sc.RedisClient.Close()
		logx.Info("Redis Client closed")
	}

	// if sc.DB != nil {
	// 	err = sc.DB.Close()
	// 	logx.Info("Database closed")
	// }

	return err
}
