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

// AdvanceTx 在当前事务中推进一条已持久化事件的消费进度, 仅允许从 expectedOffset 推进到 expectedOffset + 1
func (d *MemberWalletCheckpoint) AdvanceTx(ctx context.Context, tx *ent.Tx, topic string, partition int32, expectedOffset int64) error {
	if err := validateCheckpoint(topic, partition); err != nil {
		return err
	}
	if expectedOffset < 0 || expectedOffset == math.MaxInt64 {
		return fmt.Errorf("%w: expected_offset=%d", ErrInvalidCheckpoint, expectedOffset)
	}

	// 使用条件更新防止错误进度覆盖, 不允许跳过中间 Offset
	affected, err := tx.MemberWalletCheckpoint.Update().
		Where(
			memberwalletcheckpoint.TopicEQ(topic),               // Kafka 资金事件 Topic
			memberwalletcheckpoint.PartitionEQ(partition),       // Kafka Partition 编号
			memberwalletcheckpoint.NextOffsetEQ(expectedOffset), // 下一条需要持久化的 Kafka Offset
		).
		SetNextOffset(expectedOffset + 1).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("advance checkpoint: topic=%s partition=%d offset=%d: %w", topic, partition, expectedOffset, err)
	}

	if affected != 1 {
		return fmt.Errorf("%w: topic=%s partition=%d expected_offset=%d", ErrCheckpointNotAdvanced, topic, partition, expectedOffset)
	}

	return nil
}

// validateCheckpoint 校验 Kafka Topic 和 Partition
func validateCheckpoint(topic string, partition int32) error {
	if topic == "" || len(topic) > 249 || partition < 0 {
		return fmt.Errorf("%w: topic=%q partition=%d", ErrInvalidCheckpoint, topic, partition)
	}

	return nil
}
