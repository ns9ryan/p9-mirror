package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func testRedisCache(t *testing.T) *RedisCache {
	t.Helper()
	s := miniredis.RunT(t)
	c, err := NewRedisCache(RedisConf{Host: s.Addr(), KeyPrefix: "ipgeo"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// 同 IP 不同源
func TestRedisCache_SameIPDifferentSources(t *testing.T) {
	c := testRedisCache(t)
	ctx := context.Background()
	a := sampleRecord("8.8.8.8", "maxmind", "US")
	a.ExpiresAt = time.Now().Add(time.Hour)
	b := sampleRecord("8.8.8.8", "ip2region", "CN")
	b.ExpiresAt = time.Now().Add(time.Hour)
	if err := c.Set(ctx, a); err != nil {
		t.Fatal(err)
	}
	if err := c.Set(ctx, b); err != nil {
		t.Fatal(err)
	}
	gotA, err := c.Get(ctx, Key("maxmind", "8.8.8.8"))
	if err != nil || gotA == nil || gotA.CountryCode != "US" {
		t.Fatalf("a=%+v err=%v", gotA, err)
	}
	gotB, err := c.Get(ctx, Key("ip2region", "8.8.8.8"))
	if err != nil || gotB == nil || gotB.CountryCode != "CN" {
		t.Fatalf("b=%+v err=%v", gotB, err)
	}
	if err := c.DeleteByIP(ctx, "8.8.8.8"); err != nil {
		t.Fatal(err)
	}
	gotA, _ = c.Get(ctx, Key("maxmind", "8.8.8.8"))
	gotB, _ = c.Get(ctx, Key("ip2region", "8.8.8.8"))
	if gotA != nil || gotB != nil {
		t.Fatalf("not deleted a=%+v b=%+v", gotA, gotB)
	}
}

// 续期用 PTTL
func TestRedisCache_TouchUsesPTTL(t *testing.T) {
	c := testRedisCache(t)
	ctx := context.Background()
	rec := sampleRecord("1.1.1.1", "maxmind", "US")
	rec.ExpiresAt = time.Now().Add(time.Minute)
	if err := c.Set(ctx, rec); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Hour)
	if err := c.Touch(ctx, Key("maxmind", "1.1.1.1"), later); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, Key("maxmind", "1.1.1.1"))
	if err != nil || got == nil {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if !got.ExpiresAt.After(time.Now().Add(50 * time.Minute)) {
		t.Fatalf("ttl not slid: %v", got.ExpiresAt)
	}
}

func TestRedisCache_miss(t *testing.T) {
	c := testRedisCache(t)
	got, err := c.Get(context.Background(), Key("maxmind", "9.9.9.9"))
	if err != nil || got != nil {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
