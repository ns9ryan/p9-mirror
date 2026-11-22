package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"

	"oa.98ent.com/p9/operator-wallet/pkg/database"
)

type Config struct {
	zrpc.RpcServerConf

	// 数据库配置
	DatabaseConf database.DatabaseConf

	// 钱包运行时配置
	WalletRuntimeConf WalletRuntimeConf
}

// WalletRuntimeConf 钱包 Actor 运行时配置
type WalletRuntimeConf struct {
	// Kafka 配置
	KafkaBrokers       []string // Kafka Broker 地址
	KafkaTopic         string   // 资金事件 Topic
	KafkaConsumerGroup string   // Consumer Group ID

	// 并发配置
	ShardCount      int // Shard 数量, 必须与 Topic Partition 数量一致
	WorkerCount     int // 共享 Worker 数量
	MailboxCapacity int // 每个 Actor 的 Mailbox 容量

	PublishTimeout time.Duration // 单次 Kafka 写入超时时间
}

// IsDebug 判断是否为开发或测试模式
func (c Config) IsDebug() bool {
	return c.Mode == service.DevMode || c.Mode == service.TestMode
}
