package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	ErrKafkaProducerUnavailable = errors.New("wallet kafka producer unavailable")    // Kafka生产者不可用
	ErrKafkaInvalidRecord       = errors.New("invalid wallet kafka record")          // 资金事件消息无效
	ErrKafkaWriteUncertain      = errors.New("wallet kafka write outcome uncertain") // Kafka写入结果不明确
)

// eventProducer 定义资金事件的可靠发送接口
type eventProducer interface {
	publish(ctx context.Context, partition int32, event walletEvent) error
}

// kafkaProducer 负责将完整资金事件写入指定 Kafka Partition
type kafkaProducer struct {
	client *kgo.Client // Kafka客户端
	topic  string      // 资金日志Topic
}

// newKafkaProducer 创建钱包资金事件生产者
func newKafkaProducer(brokers []string, topic string) (*kafkaProducer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("%w: empty brokers", ErrKafkaProducerUnavailable)
	}
	for _, broker := range brokers {
		if broker == "" || strings.TrimSpace(broker) != broker {
			return nil, fmt.Errorf("%w: invalid broker address", ErrKafkaProducerUnavailable)
		}
	}
	if topic == "" || strings.TrimSpace(topic) != topic {
		return nil, fmt.Errorf("%w: invalid topic", ErrKafkaProducerUnavailable)
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),                    // 配置Kafka Broker初始连接地址
		kgo.RecordPartitioner(kgo.ManualPartitioner()), // 使用消息指定的Partition, 不自动计算分区
		kgo.RequiredAcks(kgo.AllISRAcks()),             // 等待当前所有ISR副本确认写入
	)
	if err != nil {
		return nil, fmt.Errorf("create wallet kafka producer: %w", err)
	}

	// franz-go默认启用幂等生产, 不关闭该能力
	return &kafkaProducer{
		client: client,
		topic:  topic,
	}, nil
}

// publish 写入完整资金事件, 仅在 Kafka 可靠确认后返回成功
// ctx 应由钱包服务控制, 不能直接使用可能提前取消的RPC Context
func (p *kafkaProducer) publish(ctx context.Context, partition int32, event walletEvent) error {
	if p == nil || p.client == nil {
		return ErrKafkaProducerUnavailable
	}
	if ctx == nil {
		return fmt.Errorf("%w: nil context", ErrKafkaProducerUnavailable)
	}

	record, err := p.buildRecord(partition, event)
	if err != nil {
		return err
	}

	// 未获得 Kafka 成功确认时, 无法判断消息是否已经写入
	// 调用方必须保留原始事件及 operation_id, 并将当前钱包标记为故障
	// 故障解除前, 不得继续处理该钱包的资金操作
	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("%w: %w", ErrKafkaWriteUncertain, err)
	}

	return nil
}

// buildRecord 构造指定Partition的Kafka消息, 不执行网络操作
func (p *kafkaProducer) buildRecord(partition int32, event walletEvent) (*kgo.Record, error) {
	if p == nil || p.topic == "" {
		return nil, fmt.Errorf("%w: empty topic", ErrKafkaInvalidRecord)
	}
	if partition < 0 {
		return nil, fmt.Errorf("%w: invalid partition %d", ErrKafkaInvalidRecord, partition)
	}
	if event.SchemaVersion != walletEventSchemaVersion {
		return nil, fmt.Errorf("%w: unsupported schema version", ErrKafkaInvalidRecord)
	}
	if event.WalletGlobalID == uuid.Nil {
		return nil, fmt.Errorf("%w: empty wallet global ID", ErrKafkaInvalidRecord)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal wallet event: %w", err)
	}

	return &kgo.Record{
		Topic:     p.topic,                 // 资金事件Topic
		Partition: partition,               // 指定钱包所属分区
		Key:       event.WalletGlobalID[:], // 钱包全局ID, 作为Kafka消息Key
		Value:     data,                    // 序列化后的完整资金事件
	}, nil
}

// close 关闭Kafka客户端
// 必须等待所有已接受的资金操作结束后再调用
func (p *kafkaProducer) close() {
	if p != nil && p.client != nil {
		p.client.Close()
	}
}
