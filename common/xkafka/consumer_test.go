package xkafka

import (
	"context"
	"testing"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"oa.98ent.com/p9/common/xkafka/topic"
)

func TestConsumeSkipped(t *testing.T) {
	if qs := Consume(kq.KqConf{}, nil); qs != nil {
		t.Fatalf("got=%v", qs)
	}
	if qs := Consume(kq.KqConf{}, func(context.Context, Message) error { return nil }); qs != nil {
		t.Fatalf("empty topic got=%v", qs)
	}
	StopMany(nil)
}

func TestConsumeManyNoMatch(t *testing.T) {
	if qs := ConsumeMany(nil, nil); qs != nil {
		t.Fatalf("got=%v", qs)
	}
	if qs := ConsumeMany(nil, map[string]Handler{
		topic.GameUserRegister: func(context.Context, Message) error { return nil },
	}); len(qs) != 0 {
		t.Fatalf("skip without conf, got=%d", len(qs))
	}
	if qs := ConsumeMany([]kq.KqConf{{Topic: topic.AgentCreate}}, nil); qs != nil {
		t.Fatalf("got=%v", qs)
	}
}

func TestStopEmptyQueue(t *testing.T) {
	q := &consumerQueue{conf: kq.KqConf{Topic: "t", Group: "g"}}
	done := make(chan struct{})
	go func() {
		q.Stop()
		q.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop hung on empty queue")
	}
}

func TestTopicConstants(t *testing.T) {
	if topic.GameUserRegister == "" || topic.GameUserDepositOrderApproved == "" {
		t.Fatal("empty topic")
	}
	if topic.PartnerPromotionOrderPayout == "" || topic.AgentCreate == "" {
		t.Fatal("empty topic")
	}
}
