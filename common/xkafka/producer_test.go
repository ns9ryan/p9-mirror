package xkafka

import (
	"context"
	"testing"

	"oa.98ent.com/p9/common/xkafka/topic"
)

func TestProducerEmptyBrokersNoop(t *testing.T) {
	p := NewProducerWithDef(nil)
	ctx := context.Background()
	if err := p.Push(ctx, topic.GameUserRegister, `{}`); err != nil {
		t.Fatal(err)
	}
	if err := p.PushWithKey(ctx, topic.GameUserRegister, "k", `{}`); err != nil {
		t.Fatal(err)
	}
	payload, err := NewPayload(topic.GameUserRegister, map[string]int{"id": 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.PushPayload(ctx, topic.GameUserRegister, payload); err != nil {
		t.Fatal(err)
	}
	if err := p.PushPayloadWithKey(ctx, topic.GameUserRegister, "k", payload); err != nil {
		t.Fatal(err)
	}
	p.Close()
	(*Producer)(nil).Close()
}
