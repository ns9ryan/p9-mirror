package xkafka

import (
	"encoding/json"
	"testing"

	"oa.98ent.com/p9/common/xkafka/topic"
)

type sampleData struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

func TestNewPayloadRoundTrip(t *testing.T) {
	in := sampleData{UserID: 42, Name: "alice"}
	p, err := NewPayload(topic.GameUserRegister, in)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != topic.GameUserRegister {
		t.Fatalf("name=%s", p.Name)
	}
	if p.V != DefaultPayloadVersion {
		t.Fatalf("v=%d", p.V)
	}
	if p.ID == "" {
		t.Fatal("id")
	}
	if p.Timestamp <= 0 {
		t.Fatal("time")
	}
	var out sampleData
	if err := p.UnmarshalData(&out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("data=%+v", out)
	}

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodePayload(raw)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.ID != p.ID || decoded.Name != p.Name || decoded.V != p.V {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func TestNewPayloadExt(t *testing.T) {
	p, err := NewPayloadExt("t", map[string]int{"n": 1}, map[string]string{"trace": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	var ext map[string]string
	if err := p.UnmarshalExt(&ext); err != nil {
		t.Fatal(err)
	}
	if ext["trace"] != "abc" {
		t.Fatalf("ext=%v", ext)
	}
}

func TestDecodePayloadEmpty(t *testing.T) {
	if _, err := DecodePayload(nil); err == nil {
		t.Fatal("expected error")
	}
	p, err := NewPayload("t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.UnmarshalData(&sampleData{}); err == nil {
		t.Fatal("expected empty data")
	}
}
