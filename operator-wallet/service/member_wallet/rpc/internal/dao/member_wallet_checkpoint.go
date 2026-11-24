package dao

import (
	"context"
	"errors"
	"fmt"
	"math"

	"oa.98ent.com/p9/operator-wallet/ent"
	"oa.98ent.com/p9/operator-wallet/ent/memberwalletcheckpoint"
)

var (
	ErrInvalidCheckpoint     = errors.New("wallet checkpoint: invalid parameters")
	ErrCheckpointNotAdvanced = errors.New("wallet checkpoint: offset mismatch")
)

// MemberWalletCheckpoint 钱包 Kafka 消费进度数据访问
type MemberWalletCheckpoint struct {
	db *ent.Client // Ent数据库客户端
}

// NewMemberWalletCheckpoint 创建消费进度 DAO
func NewMemberWalletCheckpoint(db *ent.Client) *MemberWalletCheckpoint {
	return &MemberWalletCheckpoint{db: db}
}

// Get 查询指定 Partition 的消费进度, 记录不存在时返回 ent.NotFoundError, 由启动流程决定如何初始化
func (d *MemberWalletCheckpoint) Get(ctx context.Context, topic string, partition int32) (*ent.MemberWalletCheckpoint, error) {
	if err := validateCheckpoint(topic, partition); err != nil {
		return nil, err
	}

	return d.db.MemberWalletCheckpoint.Query().
		Where(
			memberwalletcheckpoint.TopicEQ(topic),         // Kafka 资金事件 Topic
			memberwalletcheckpoint.PartitionEQ(partition), // Kafka Partition 编号
		).
		Only(ctx)
}

// GetForUpdateTx 在当前事务中查询并锁定指定 Partition 的消费进度
func (d *MemberWalletCheckpoint) GetForUpdateTx(ctx context.Context, tx *ent.Tx, topic string, partition int32) (*ent.MemberWalletCheckpoint, error) {
	if err := validateCheckpoint(topic, partition); err != nil {
		return nil, err
	}

	return tx.MemberWalletCheckpoint.Query().
		Where(
			memberwalletcheckpoint.TopicEQ(topic),         // Kafka 资金事件 Topic
			memberwalletcheckpoint.PartitionEQ(partition), // Kafka Partition 编号
		).
		ForUpdate(). // 锁定消费进度, 防止事务期间被并发修改
		Only(ctx)
}

// CreateInitialTx 在当前事务中初始化 Partition 的起始消费进度, 调用方必须先验证 Kafka 日志范围, 不得直接假定 Offset 为 0
func (d *MemberWalletCheckpoint) CreateInitialTx(ctx context.Context, tx *ent.Tx, topic string, partition int32, nextOffset int64) error {
	if err := validateCheckpoint(topic, partition); err != nil {
		return err
	}
	if nextOffset < 0 {
		return fmt.Errorf("%w: next_offset=%d", ErrInvalidCheckpoint, nextOffset)
	}

	return tx.MemberWalletCheckpoint.Create().
		SetTopic(topic).           // Kafka 资金事件 Topic
		SetPartition(partition).   // Kafka Partition 编号
		SetNextOffset(nextOffset). // 下一条需要持久化的 Kafka Offset
		Exec(ctx)
}

// AdvanceTx 在当前事务中推进已持久化消息的消费进度, 调用方必须先确认消息有效, 且不存在未处理的资金事件
func (d *MemberWalletCheckpoint) AdvanceTx(
	ctx context.Context,
	tx *ent.Tx,
	topic string,
	partition int32,
	expectedNextOffset int64,
	processedOffset int64,
) error {
	if err := validateCheckpoint(topic, partition); err != nil {
		return err
	}
	if expectedNextOffset < 0 || processedOffset < expectedNextOffset || processedOffset == math.MaxInt64 {
		return fmt.Errorf("%w: expected_next_offset=%d processed_offset=%d", ErrInvalidCheckpoint, expectedNextOffset, processedOffset)
	}

	// 仅在当前进度匹配时推进, 防止并发覆盖
	affected, err := tx.MemberWalletCheckpoint.Update().
		Where(
			memberwalletcheckpoint.TopicEQ(topic),                   // Kafka 资金事件 Topic
			memberwalletcheckpoint.PartitionEQ(partition),           // Kafka Partition 编号
			memberwalletcheckpoint.NextOffsetEQ(expectedNextOffset), // 下一条需要持久化的 Kafka Offset
		).
		SetNextOffset(processedOffset + 1).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("advance checkpoint: topic=%s partition=%d offset=%d: %w", topic, partition, processedOffset, err)
	}

	if affected != 1 {
		return fmt.Errorf("%w: topic=%s partition=%d expected_next_offset=%d", ErrCheckpointNotAdvanced, topic, partition, expectedNextOffset)
	}

	return nil
}

// ListByTopic 查询指定 Topic 的全部消费进度
func (d *MemberWalletCheckpoint) ListByTopic(ctx context.Context, topic string) (map[int32]int64, error) {
	if err := validateCheckpoint(topic, 0); err != nil {
		return nil, err
	}

	items, err := d.db.MemberWalletCheckpoint.Query().
		Where(memberwalletcheckpoint.TopicEQ(topic)). // Kafka 资金事件 Topic
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list wallet checkpoints: topic=%s: %w", topic, err)
	}

	checkpoints := make(map[int32]int64, len(items))
	for _, item := range items {
		checkpoints[item.Partition] = item.NextOffset
	}

	return checkpoints, nil
}

// validateCheckpoint 校验 Kafka Topic 和 Partition
func validateCheckpoint(topic string, partition int32) error {
	if topic == "" || len(topic) > 249 || partition < 0 {
		return fmt.Errorf("%w: topic=%q partition=%d", ErrInvalidCheckpoint, topic, partition)
	}

	return nil
}

// InitializeEmptyTopic 为全新空 Topic 初始化全部 Partition 的消费进度, 调用前必须确认 Kafka 全部分区的日志起点和终点均为 0
func (d *MemberWalletCheckpoint) InitializeEmptyTopic(ctx context.Context, topic string, partitionCount int) error {
	if err := validateCheckpoint(topic, 0); err != nil {
		return err
	}
	if partitionCount <= 0 {
		return fmt.Errorf("%w: partition_count=%d", ErrInvalidCheckpoint, partitionCount)
	}

	return WithTx(ctx, d.db, func(tx *ent.Tx) error {
		// 已有任意进度时拒绝重新初始化, 避免覆盖消费历史
		count, err := tx.MemberWalletCheckpoint.Query().
			Where(memberwalletcheckpoint.TopicEQ(topic)). // Kafka 资金事件 Topic
			Count(ctx)
		if err != nil {
			return fmt.Errorf("count wallet checkpoints: %w", err)
		}
		if count != 0 {
			return fmt.Errorf("%w: topic=%s already has %d checkpoints", ErrInvalidCheckpoint, topic, count)
		}

		// 全部分区在同一事务中创建, 任何失败都会回滚
		for partition := range int32(partitionCount) {
			if err := d.CreateInitialTx(ctx, tx, topic, partition, 0); err != nil {
				return fmt.Errorf("initialize wallet checkpoint: partition=%d: %w", partition, err)
			}
		}

		return nil
	})
}
