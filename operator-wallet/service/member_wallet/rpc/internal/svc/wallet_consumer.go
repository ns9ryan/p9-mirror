package svc

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/zeromicro/go-zero/core/logx"

	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/config"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/dao"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet"
	"oa.98ent.com/p9/operator-wallet/service/member_wallet/rpc/internal/wallet/consumer"
)

// newWalletConsumer 根据数据库消费进度创建 Kafka Consumer, 全新空 Topic 自动初始化 Checkpoint
func newWalletConsumer(ctx context.Context, conf config.WalletRuntimeConf, checkpointDAO *dao.MemberWalletCheckpoint) (*consumer.Consumer, error) {
	if conf.ShardCount <= 0 {
		return nil, fmt.Errorf("invalid wallet shard count: %d", conf.ShardCount)
	}

	// 读取当前 Topic 的全部消费进度
	checkpoints, err := checkpointDAO.ListByTopic(ctx, conf.KafkaTopic)
	if err != nil {
		return nil, err
	}

	// 仅在不存在任何消费进度时尝试首次初始化
	if len(checkpoints) == 0 {
		checkpoints, err = initializeWalletCheckpoints(ctx, conf, checkpointDAO)
		if err != nil {
			return nil, err
		}
	}

	// 消费进度必须覆盖生产端的全部 Shard
	if len(checkpoints) != conf.ShardCount {
		return nil, fmt.Errorf("wallet checkpoint count mismatch: shards=%d checkpoints=%d", conf.ShardCount, len(checkpoints))
	}

	// 校验 Kafka 实际分区及日志范围, 从数据库消费进度创建 Consumer
	return consumer.NewFromCheckpoints(ctx, conf.KafkaBrokers, conf.KafkaTopic, checkpoints)
}

// initializeWalletCheckpoints 为全新空 Topic 初始化全部分区的消费进度
func initializeWalletCheckpoints(ctx context.Context, conf config.WalletRuntimeConf, checkpointDAO *dao.MemberWalletCheckpoint) (map[int32]int64, error) {
	// 临时客户端仅用于检查 Kafka Topic
	client, err := kgo.NewClient(kgo.SeedBrokers(conf.KafkaBrokers...))
	if err != nil {
		return nil, fmt.Errorf("create Kafka metadata client: %w", err)
	}
	defer client.Close()

	// 首次初始化要求全部分区的日志起点和终点均为 0
	partitionCount, err := consumer.ValidateEmptyTopic(ctx, client, conf.KafkaTopic)
	if err != nil {
		return nil, err
	}

	// 初始化前确认 Kafka 分区数量与生产端 Shard 数量一致
	if partitionCount != conf.ShardCount {
		return nil, fmt.Errorf("wallet partition count mismatch: shards=%d partitions=%d", conf.ShardCount, partitionCount)
	}

	// 在同一数据库事务中初始化全部分区
	if err := checkpointDAO.InitializeEmptyTopic(ctx, conf.KafkaTopic, partitionCount); err != nil {
		return nil, err
	}

	// 重新读取已提交的消费进度, 以数据库实际记录为准
	return checkpointDAO.ListByTopic(ctx, conf.KafkaTopic)
}

// watchWalletConsumer 等待消费循环退出, 停止 Engine 接收新的资金操作
func watchWalletConsumer(c *consumer.Consumer, engine *wallet.Engine) {
	err := c.Wait(context.Background())

	// Consumer 退出后不再允许接收新的资金操作
	engine.StopAccepting()

	if err != nil {
		logx.Errorf("wallet consumer stopped with error: %v", err)
	}
}
