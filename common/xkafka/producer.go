package xkafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/zeromicro/go-queue/kq"
)

// Producer 按 topic 懒创建 kq.Pusher 的 Kafka 生产者。
// brokers 为空时所有 Push* 直接成功返回（no-op），便于本地未配 Kafka 时不挡主流程。
// 同一 Producer 可向多个 topic 发；每个 topic 一个 Pusher，并发 Push 安全。
// 进程退出前应 Close，否则底层 Writer 的缓冲可能未刷出。
type Producer struct {
	brokers []string
	opts    []kq.PushOption
	mu      sync.Mutex
	pushers map[string]*kq.Pusher
}

// NewProducerWithDef 用默认选项创建生产者。
// 默认：允许自动建 topic、FlushInterval=10ms。
// 不使用 SyncPush：kafka-go 默认 BatchTimeout 约 1s，同步推会把每次发送卡住约一秒。
func NewProducerWithDef(brokers []string) *Producer {
	return &Producer{
		brokers: brokers,
		opts: []kq.PushOption{
			kq.WithAllowAutoTopicCreation(),
			kq.WithFlushInterval(10 * time.Millisecond),
		},
		pushers: make(map[string]*kq.Pusher),
	}
}

// NewProducer 用自定义 kq.PushOption 创建生产者。
// opts 为空时与 NewProducer 相同。传入 opts 则完全覆盖默认，不再自动附加 FlushInterval。
func NewProducer(brokers []string, opts []kq.PushOption) *Producer {
	// 如果 opts 为空，则使用默认选项
	if len(opts) == 0 {
		return NewProducerWithDef(brokers)
	}
	// 否则使用 opts 创建生产者
	return &Producer{
		brokers: brokers,
		opts:    opts,
		pushers: make(map[string]*kq.Pusher),
	}
}

// pusher 按 topic 返回已有 Pusher，没有则创建并缓存。
// Producer 为 nil、brokers 为空或 topic 为空时返回 nil，调用方按 no-op 处理。
func (p *Producer) pusher(topic string) *kq.Pusher {
	if p == nil || len(p.brokers) == 0 || topic == "" {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.pushers[topic]; ok {
		return w
	}
	w := kq.NewPusher(p.brokers, topic, p.opts...)
	p.pushers[topic] = w
	return w
}

// Push 向 topic 发送原始字符串，不带 Key，分区由 Kafka 默认策略决定。
// 无可用 Pusher 时返回 nil。ctx 可用于超时/取消。
func (p *Producer) Push(ctx context.Context, topic, payload string) error {
	w := p.pusher(topic)
	if w == nil {
		return nil
	}
	return w.Push(ctx, payload)
}

// PushWithKey 向 topic 发送带 Key 的原始字符串。
// 相同 Key 会进同一分区，适合要保序的业务。无可用 Pusher 时返回 nil。
func (p *Producer) PushWithKey(ctx context.Context, topic, key, payload string) error {
	w := p.pusher(topic)
	if w == nil {
		return nil
	}
	return w.PushWithKey(ctx, key, payload)
}

// PushPayload 将 Payload 序列化为 JSON 后发送，不带 Key。
// 信封字段见 Payload（name/data/time/v/ext）。Marshal 失败才返回 error。
func (p *Producer) PushPayload(ctx context.Context, topic string, payload Payload) error {
	return p.PushPayloadWithKey(ctx, topic, "", payload)
}

// PushPayloadWithKey 将 Payload 序列化为 JSON 后发送。
// key 为空等价于 PushPayload；非空则走 PushWithKey 以固定分区。
func (p *Producer) PushPayloadWithKey(ctx context.Context, topic, key string, payload Payload) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if key == "" {
		return p.Push(ctx, topic, string(raw))
	}
	return p.PushWithKey(ctx, topic, key, string(raw))
}

// Close 关闭并丢弃已创建的全部 Pusher，刷新剩余缓冲。
// Producer 为 nil 时直接返回。关闭后同一实例再 Push 会为对应 topic 新建 Pusher。
func (p *Producer) Close() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for topic, w := range p.pushers {
		if w != nil {
			_ = w.Close()
		}
		delete(p.pushers, topic)
	}
}
