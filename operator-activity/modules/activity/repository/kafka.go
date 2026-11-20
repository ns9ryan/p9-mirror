package repository

import (
	"context"
	"encoding/json"
	"time"

	"oa.98ent.com/p9/common/xkafka/topic"

	"oa.98ent.com/p9/operator-activity/modules/activity/types"

	"oa.98ent.com/p9/common/xkafka"

	"github.com/zeromicro/go-queue/kq"
)

// KafkaPublisher 基于 go-queue/kq 的消息发布
type KafkaPublisher struct {
	pusher *xkafka.Producer
}

// NewKafkaPublisher 创建发布器；brokers 为空时发布为 no-op（便于单测）
// 不使用 SyncPush：kafka-go Writer 默认 BatchTimeout=1s，同步单条推送会卡住约 1 秒。
// 走 kq ChunkExecutor 异步批量，FlushInterval 压低投递延迟。
func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	if len(brokers) == 0 {
		return nil
	}
	opts := []kq.PushOption{
		// kq.WithSyncPush(),
		kq.WithAllowAutoTopicCreation(),
		kq.WithFlushInterval(10 * time.Millisecond),
		// kq.WithChunkSize(100),
	}
	return &KafkaPublisher{
		pusher: xkafka.NewProducer(brokers, opts),
	}
}

func (k *KafkaPublisher) Close() {
	if k.pusher != nil {
		k.pusher.Close()
	}
}

func (k *KafkaPublisher) PublishPayout(ctx context.Context, order *types.ActivityOrder) error {
	if k.pusher == nil {
		return nil
	}
	// 业务数据JSON字符串
	dataRaw := `{"order_id":` + itoa(order.ID) + `,"activity_id":` + itoa(order.ActivityID) +
		`,"member_id":` + itoa(order.MemberID) + `,"amount":` + itoa(order.RewardsTotal) + `}`

	payload, err := xkafka.NewPayload(topic.PartnerPromotionOrderPayout, json.RawMessage(dataRaw))
	if err != nil {
		return err
	}
	return k.pusher.PushPayload(ctx, topic.PartnerPromotionOrderPayout, payload)
}

func (k *KafkaPublisher) PublishOrderCreated(ctx context.Context, orders []*types.ActivityOrder) error {
	if k.pusher == nil {
		return nil
	}
	for _, order := range orders {
		// 业务数据JSON字符串
		dataRaw := `{"order_id":` + itoa(order.ID) + `,"activity_id":` + itoa(order.ActivityID) +
			`,"member_id":` + itoa(order.MemberID) + `}`
		// 创建xkafka.Payload
		payload, err := xkafka.NewPayload(topic.PartnerPromotionOrderCreated, json.RawMessage(dataRaw))
		if err != nil {
			return err
		}
		if err := k.pusher.PushPayload(ctx, topic.PartnerPromotionOrderCreated, payload); err != nil {
			return err
		}
	}
	return nil
}

// PublishDepositApproved 发布充值审核通过消息
func (k *KafkaPublisher) PublishDepositApproved(ctx context.Context, data *types.DepositPayload) error {
	if k.pusher == nil || data == nil {
		return nil
	}
	// 序列化业务数据
	dataRaw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 创建xkafka.Payload
	payload, err := xkafka.NewPayload(topic.GameUserDepositOrderApproved, dataRaw)
	if err != nil {
		return err
	}
	return k.pusher.PushPayload(ctx, topic.GameUserDepositOrderApproved, payload)
}

// PublishRegister 发布会员注册消息
func (k *KafkaPublisher) PublishRegister(ctx context.Context, data *types.RegisterPayload) error {
	if k.pusher == nil || data == nil {
		return nil
	}
	// 业务数据
	dataRaw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 创建xkafka.Payload
	payload, err := xkafka.NewPayload(topic.GameUserRegister, dataRaw)
	if err != nil {
		return err
	}
	return k.pusher.PushPayload(ctx, topic.GameUserRegister, payload)
}

// PublishBetSettled 发布投注结算消息
func (k *KafkaPublisher) PublishBetSettled(ctx context.Context, data *types.BetPayload) error {
	if k.pusher == nil || data == nil {
		return nil
	}
	// 业务数据
	dataRaw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 创建xkafka.Payload
	payload, err := xkafka.NewPayload(topic.GameUserGameOrderSettled, dataRaw)
	if err != nil {
		return err
	}
	return k.pusher.PushPayload(ctx, topic.GameUserGameOrderSettled, payload)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
