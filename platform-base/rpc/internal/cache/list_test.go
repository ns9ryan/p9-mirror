package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"

	"oa.98ent.com/p9/platform-base/rpc/ent"
)

func TestListRedisRoundTrip(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	calls := 0
	source := &ent.Timezone{ID: 1, Status: 1, Code: "Asia/Shanghai", UpdatedAt: now}
	list := NewList(client, "timezone", 24*time.Hour, func(context.Context) ([]*ent.Timezone, error) {
		calls++
		return []*ent.Timezone{source}, nil
	})

	check := func(wantStatus int64, wantCalls int) {
		t.Helper()
		rows, err := list.All(ctx)
		if err != nil || len(rows) != 1 || rows[0].Status != wantStatus ||
			rows[0].Code != source.Code || !rows[0].UpdatedAt.Equal(now) || calls != wantCalls {
			t.Fatalf("rows=%+v calls=%d err=%v", rows, calls, err)
		}
	}

	check(1, 1)
	key := "platform-base:timezone:list:msgpack:v1:0"
	payload, err := server.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	var cached []*ent.Timezone
	if err := msgpack.Unmarshal([]byte(payload), &cached); err != nil || len(cached) != 1 || cached[0].ID != source.ID {
		t.Fatalf("expected raw MessagePack payload in Redis, err=%v", err)
	}

	if ttl := server.TTL(key); ttl < 24*time.Hour || ttl >= 24*time.Hour+5*time.Minute {
		t.Fatalf("unexpected cache TTL: %v", ttl)
	}

	source.Status = 2
	check(1, 1)
	list.Invalidate(ctx)
	check(2, 2)

	key = "platform-base:timezone:list:msgpack:v1:1"
	if err := server.Set(key, "broken payload"); err != nil {
		t.Fatal(err)
	}
	check(2, 3)
	check(2, 3)
}
