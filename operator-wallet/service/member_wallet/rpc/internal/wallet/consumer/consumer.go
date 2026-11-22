package consumer

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	errEmptyBrokers = errors.New("wallet consumer: empty brokers")
	errEmptyTopic   = errors.New("wallet consumer: empty topic")
	errEmptyGroupID = errors.New("wallet consumer: empty group ID")
)

// Config 配置钱包资金事件消费者
type Config struct {
	Brokers []string // Kafka Broker 地址
	Topic   string   // 资金事件 Topic
	GroupID string   // Consumer Group ID
}

// Consumer 负责消费钱包资金事件
type Consumer struct {
	client *kgo.Client // Kafka 消费客户端

	closeOnce sync.Once // 确保客户端只关闭一次
}

// New 创建钱包资金事件消费者
func New(config Config) (*Consumer, error) {
	if len(config.Brokers) == 0 {
		return nil, errEmptyBrokers
	}

	for _, broker := range config.Brokers {
		if broker == "" || strings.TrimSpace(broker) != broker {
			return nil, fmt.Errorf("wallet consumer: invalid broker address")
		}
	}

	if config.Topic == "" || strings.TrimSpace(config.Topic) != config.Topic {
		return nil, errEmptyTopic
	}

	if config.GroupID == "" || strings.TrimSpace(config.GroupID) != config.GroupID {
		return nil, errEmptyGroupID
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumerGroup(config.GroupID),
		kgo.ConsumeTopics(config.Topic),
		kgo.DisableAutoCommit(), // 由数据库事务成功后手动提交 Offset
	)
	if err != nil {
		return nil, fmt.Errorf("create wallet consumer: %w", err)
	}

	return &Consumer{
		client: client,
	}, nil
}

// Close 关闭 Kafka 消费客户端
func (c *Consumer) Close() {
	if c == nil {
		return
	}

	c.closeOnce.Do(func() {
		if c.client != nil {
			c.client.Close()
		}
	})
}
