package xkafka

import (
	"context"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
)

const (
	// dialTimeout 建立 broker 连接的超时。避免 Kafka 不可达时 Reader 长期阻塞。
	dialTimeout = 3 * time.Second
	// stopTimeout 单消费者 Close/Wait 以及 StopMany 汇总等待的上限。
	// kafka-go Reader.Close 在组协调超时或 broker 不可达时可能卡住，超时后放弃等待以便进程退出。
	stopTimeout = 3 * time.Second
)

// Message 消费到的一条 Kafka 记录（相对 kafka-go.Message 的精简视图）。
// Topic/Partition/Offset 用于日志与排障；Key/Value 为消息体。
// 不包含 HighWaterMark：需要积压监控时再从底层 FetchMessage 透传。
type Message struct {
	Topic     string // 实际消费到的 topic；空则回落到配置中的 Topic
	Partition int    // 分区号
	Offset    int64  // 分区内偏移量
	Key       string // 消息 Key，可用于分区路由；无 Key 时为空
	Value     string // 消息 Value，业务侧通常再 DecodePayload
}

// Handler 消费回调。
// 返回 error 且 conf.ForceCommit=false 时本条不提交 offset，下次会再次投递（at-least-once）。
// 返回 nil（或 ForceCommit=true）后才会 CommitMessages。
// ctx 在 Stop 时取消，回调应尽快返回，避免拖住退出。
type Handler func(ctx context.Context, msg Message) error

// consumerQueue 实现 queue.MessageQueue，一个实例对应一个 topic + group。
// Start 按 conf.Consumers 拉起多个 kafka.Reader 并行 Fetch；Stop 只执行一次。
type consumerQueue struct {
	conf     kq.KqConf
	handler  Handler
	mu       sync.Mutex
	cancel   context.CancelFunc
	readers  []*kafka.Reader
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// Start 同步创建 Reader 并启动拉取循环，满足 queue.MessageQueue。
// Consumers<=0 时按 1 个 Reader；Offset=="first" 从最早消息读，否则从最新。
// CommitInterval=0 表示手动提交，与 ForceCommit 配合。
// 应在 Consume/startQueue 内调用一次；重复调用会再追加 Reader，不作为公开重启接口。
func (q *consumerQueue) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	q.mu.Lock()
	q.cancel = cancel
	q.mu.Unlock()

	n := max(q.conf.Consumers, 1)
	offset := kafka.LastOffset
	if q.conf.Offset == "first" {
		offset = kafka.FirstOffset
	}
	minBytes := q.conf.MinBytes
	if minBytes <= 0 {
		minBytes = 10240
	}
	maxBytes := q.conf.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 10485760
	}

	for range n {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: q.conf.Brokers,
			GroupID: q.conf.Group,
			Topic:   q.conf.Topic,
			Dialer: &kafka.Dialer{
				Timeout:   dialTimeout,
				DualStack: true,
			},
			StartOffset:    offset,
			MinBytes:       minBytes,
			MaxBytes:       maxBytes,
			MaxWait:        time.Second,
			CommitInterval: 0, // 手动提交，配合 ForceCommit
		})
		q.mu.Lock()
		q.readers = append(q.readers, r)
		q.mu.Unlock()
		q.wg.Add(1)
		go q.loop(ctx, r)
	}
}

// loop 单 Reader 拉取循环：FetchMessage → Handler → 按结果决定是否提交 offset。
// ctx 取消（Stop）或 Fetch/Commit 因 ctx 失败时退出；其它 Fetch 错误打日志后继续拉，避免单次失败打死消费者。
func (q *consumerQueue) loop(ctx context.Context, r *kafka.Reader) {
	defer q.wg.Done()
	for {
		km, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logx.Errorf("[xkafka] fetch message error topic=%s err=%v", q.conf.Topic, err)
			continue
		}
		topicName := km.Topic
		if topicName == "" {
			topicName = q.conf.Topic
		}
		msg := Message{
			Topic:     topicName,
			Partition: km.Partition,
			Offset:    km.Offset,
			Key:       string(km.Key),
			Value:     string(km.Value),
		}
		if herr := q.handler(ctx, msg); herr != nil && !q.conf.ForceCommit {
			logx.Errorf("[xkafka] consume error skip commit topic=%s partition=%d offset=%d err=%v",
				msg.Topic, msg.Partition, msg.Offset, herr)
			continue
		}
		if err := r.CommitMessages(ctx, km); err != nil {
			if ctx.Err() != nil {
				return
			}
			logx.Errorf("[xkafka] commit failed topic=%s partition=%d offset=%d err=%v",
				msg.Topic, msg.Partition, msg.Offset, err)
		}
	}
}

// Stop 停止本 topic 的全部 Reader，满足 queue.MessageQueue。
// 只生效一次：取消 Fetch 用的 ctx，再异步 Close Reader 并等待 loop 退出。
// Close 若超过 stopTimeout 未返回则放弃等待并打错误日志，避免 Ctrl+C 卡死进程。
func (q *consumerQueue) Stop() {
	q.stopOnce.Do(func() {
		q.mu.Lock()
		cancel := q.cancel
		readers := append([]*kafka.Reader(nil), q.readers...)
		q.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		done := make(chan struct{})
		go func() {
			for _, r := range readers {
				_ = r.Close()
			}
			q.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(stopTimeout):
			logx.Errorf("[xkafka] stop timeout topic=%s group=%s", q.conf.Topic, q.conf.Group)
		}
	})
}

// startQueue 构造并立即 Start 一个 consumerQueue。
// 同步启动是为了 Consume 返回时 Reader 已就绪，避免 Stop 抢在 go Start() 之前导致泄漏。
func startQueue(conf kq.KqConf, handler Handler) queue.MessageQueue {
	q := &consumerQueue{conf: conf, handler: handler}
	q.Start()
	logx.Infof("[xkafka] consumer started topic=%s group=%s forceCommit=%v",
		conf.Topic, conf.Group, conf.ForceCommit)
	return q
}

// Consume 按 conf 启动单个 topic 的消费者并开始拉取。
// handler 为 nil 或 Topic 为空时跳过，返回 nil。
// 返回的 MessageQueue 交给 Stop/StopMany；进程退出前必须停，否则 Reader 会一直占着消费组。
func Consume(conf kq.KqConf, handler Handler) queue.MessageQueue {
	if handler == nil || conf.Topic == "" {
		logx.Error("[xkafka] consume skipped: empty topic or nil handler")
		return nil
	}
	return startQueue(conf, handler)
}

// ConsumeMany 按 handlers 的 topic 名匹配 confs，只启动「有配置且已注册回调」的消费者。
// handlers 为空直接返回 nil；某 topic 有 handler 无 conf 则跳过并打错误；有 conf 无 handler 仅 info 忽略。
// 同一 topic 在 confs 中出现多次时后者覆盖前者。返回值交给 StopMany。
func ConsumeMany(confs []kq.KqConf, handlers map[string]Handler) []queue.MessageQueue {
	if len(handlers) == 0 {
		return nil
	}
	confByTopic := make(map[string]kq.KqConf, len(confs))
	for _, conf := range confs {
		if conf.Topic == "" {
			continue
		}
		confByTopic[conf.Topic] = conf
	}

	queues := make([]queue.MessageQueue, 0, len(handlers))
	for topicName, h := range handlers {
		if h == nil {
			continue
		}
		conf, ok := confByTopic[topicName]
		if !ok {
			logx.Errorf("[xkafka] handler has no kafka config, skip topic=%s", topicName)
			continue
		}
		queues = append(queues, startQueue(conf, h))
	}

	for topicName := range confByTopic {
		if _, ok := handlers[topicName]; !ok {
			logx.Infof("[xkafka] config topic not registered, ignore topic=%s", topicName)
		}
	}
	return queues
}

// Stop 停止单个消费者。q 为 nil 时仅打日志。内部 Stop 幂等。
func Stop(q queue.MessageQueue) {
	if q != nil {
		q.Stop()
	}
	logx.Info("[xkafka] consumer stopped")
}

// StopMany 并行停止多个消费者，总等待不超过 stopTimeout。
// 超时后仍返回，后台 Close 可能还在跑；进程即将退出时这样可避免 defer 卡死。
// nil 项跳过。
func StopMany(queues []queue.MessageQueue) {
	var wg sync.WaitGroup
	for _, q := range queues {
		if q == nil {
			continue
		}
		wg.Add(1)
		go func(q queue.MessageQueue) {
			defer wg.Done()
			q.Stop()
		}(q)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(stopTimeout):
		logx.Error("[xkafka] stop many timeout")
	}
	logx.Info("[xkafka] all consumers stopped")
}
