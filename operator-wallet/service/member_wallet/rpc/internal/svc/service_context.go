package svc

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/migrate"
	_ "oa.98ent.com/p9/operator-wallet/ent/runtime"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/config"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet/consumer"
)

// ServiceContext 会员钱包RPC服务上下文
type ServiceContext struct {
	Config config.Config
	DB     *ent.Client // Ent数据库客户端
	DAO    *dao.Group  // 钱包数据访问

	// 钱包运行时
	WalletEngine   *wallet.Engine     // 钱包 Actor 运行时
	WalletConsumer *consumer.Consumer // Kafka 资金事件消费者
	walletCancel   context.CancelFunc // 钱包服务生命周期取消函数

	// 钱包标识缓存
	walletIndexMu sync.RWMutex                        // 保护钱包标识缓存
	walletIndex   map[walletIndexKey]walletIndexEntry // 厅和会员到钱包标识的映射

	// 关闭状态
	closeOnce sync.Once
	closeErr  error
}

// NewServiceContext 创建会员钱包RPC服务上下文
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 创建数据库驱动
	driver, err := c.DatabaseConf.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("create member wallet database driver: %w", err)
	}

	// 创建Ent客户端配置
	entOpts := []ent.Option{
		ent.Log(logx.Info), // 使用go-zero日志输出SQL
		ent.Driver(driver), // 设置数据库驱动
	}

	// 开发和测试环境开启SQL调试
	if c.IsDebug() {
		entOpts = append(entOpts, ent.Debug())
	}

	// 创建Ent数据库客户端
	db := ent.NewClient(entOpts...)

	// 启动时同步数据库结构
	if err := db.Schema.Create(
		context.Background(),
		migrate.WithDropColumn(false),  // 禁止自动删除字段
		migrate.WithDropIndex(false),   // 禁止自动删除索引
		migrate.WithForeignKeys(false), // 不创建数据库外键约束
	); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sync member wallet database schema: %w", err)
	}

	// 创建钱包服务生命周期Context
	walletCtx, walletCancel := context.WithCancel(context.Background())

	// 创建钱包Engine及Kafka Producer
	walletConf := c.WalletRuntimeConf
	walletEngine, err := wallet.NewEngine(walletCtx, wallet.EngineConfig{
		KafkaBrokers:    walletConf.KafkaBrokers,    // Kafka Broker 地址
		KafkaTopic:      walletConf.KafkaTopic,      // 资金事件 Topic
		ShardCount:      walletConf.ShardCount,      // Shard 数量, 必须与 Topic Partition 数量一致
		WorkerCount:     walletConf.WorkerCount,     // 共享 Worker 数量
		MailboxCapacity: walletConf.MailboxCapacity, // 每个 Actor 的 Mailbox 容量
		PublishTimeout:  walletConf.PublishTimeout,  // 单次 Kafka 写入超时时间
	})
	if err != nil {
		walletCancel()
		_ = db.Close()
		return nil, fmt.Errorf("create member wallet engine: %w", err)
	}

	// 创建Kafka Consumer
	// walletConsumer, err := consumer.New(consumer.Config{
	// 	Brokers: walletConf.KafkaBrokers,       // Kafka Broker 地址
	// 	Topic:   walletConf.KafkaTopic,         // 资金事件 Topic
	// 	GroupID: walletConf.KafkaConsumerGroup, // Consumer Group ID
	// })
	// if err != nil {
	// 	_ = walletEngine.Close(context.Background())
	// 	walletCancel()
	// 	_ = db.Close()
	// 	return nil, fmt.Errorf("create member wallet consumer: %w", err)
	// }

	return &ServiceContext{
		Config:       c,
		DB:           db,
		DAO:          dao.New(db),
		WalletEngine: walletEngine,
		// WalletConsumer: walletConsumer,
		walletCancel: walletCancel,

		// 初始化钱包标识缓存
		walletIndex: make(map[walletIndexKey]walletIndexEntry),
	}, nil
}

// Close 按依赖顺序关闭钱包运行时和数据库, Engine 未完全退出时返回等待错误并保留依赖, 调用方可使用新的 Context 再次调用
func (s *ServiceContext) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}

	if ctx == nil {
		return errors.New("close member wallet service context: nil context")
	}

	// 先等待全部Actor和Worker退出, 再关闭Kafka Producer
	var engineErr error
	if s.WalletEngine != nil {
		engineErr = s.WalletEngine.Close(ctx)

		// 超时或仍有任务时, 保留服务依赖供后续重试
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(engineErr, wallet.ErrWorkerPoolBusy) {
			return engineErr
		}
	}

	// Engine退出后才取消服务Context并关闭数据库
	s.closeOnce.Do(func() {
		if s.WalletConsumer != nil {
			s.WalletConsumer.Close()
		}

		if s.walletCancel != nil {
			s.walletCancel()
		}

		if s.DB != nil {
			if err := s.DB.Close(); err != nil {
				s.closeErr = fmt.Errorf("close member wallet database: %w", err)
			}
		}
	})

	return errors.Join(engineErr, s.closeErr)
}
