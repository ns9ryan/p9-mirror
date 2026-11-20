package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
)

func sampleRecord(ip, source, country string) *geo.Record {
	return &geo.Record{
		IP:           ip,
		CountryCode:  country,
		CountryNames: geo.LangNames{"en": "United States", "zh-CN": "美国"},
		Source:       source,
		UpdatedAt:    time.Now(),
	}
}

// 默认表驱动需要 DB
func TestOpenCache_DefaultTableNeedsDB(t *testing.T) {
	_, err := OpenCache("", nil, RedisConf{})
	if !errors.Is(err, geo.ErrMissingDSN) {
		t.Fatalf("got %v", err)
	}
}

func TestOpenCache_memory(t *testing.T) {
	c, err := OpenCache(CacheDriverMemory, nil, RedisConf{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(*MemoryCache); !ok {
		t.Fatalf("got %T", c)
	}
}

// redis 缺少 Host
func TestOpenCache_RedisMissingHost(t *testing.T) {
	_, err := OpenCache(CacheDriverRedis, nil, RedisConf{})
	if !errors.Is(err, geo.ErrMissingRedis) {
		t.Fatalf("got %v", err)
	}
}

// 未知驱动
func TestOpenCache_UnknownDriver(t *testing.T) {
	_, err := OpenCache("mongo", nil, RedisConf{})
	if !errors.Is(err, geo.ErrBadCacheDriver) {
		t.Fatalf("got %v", err)
	}
}

func TestEncodeDecodeRecord(t *testing.T) {
	in := sampleRecord("8.8.8.8", "maxmind", "US")
	in.ExpiresAt = time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	raw, err := encodeRecord(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeRecord(raw)
	if err != nil || got.IP != in.IP || got.Source != in.Source || got.CountryCode != in.CountryCode {
		t.Fatalf("got %+v err=%v", got, err)
	}
	if !got.ExpiresAt.Equal(in.ExpiresAt) {
		t.Fatalf("expires %v != %v", got.ExpiresAt, in.ExpiresAt)
	}
}

// 同 IP 不同源各一行
func TestMemoryCache_SameIPDifferentSources(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	_ = c.Set(ctx, sampleRecord("8.8.8.8", "maxmind", "US"))
	_ = c.Set(ctx, sampleRecord("8.8.8.8", "ip2region", "CN"))
	a, _ := c.Get(ctx, Key("maxmind", "8.8.8.8"))
	b, _ := c.Get(ctx, Key("ip2region", "8.8.8.8"))
	if a == nil || b == nil || a.CountryCode == b.CountryCode {
		t.Fatalf("a=%+v b=%+v", a, b)
	}
	_ = c.DeleteByIP(ctx, "8.8.8.8")
	a, _ = c.Get(ctx, Key("maxmind", "8.8.8.8"))
	b, _ = c.Get(ctx, Key("ip2region", "8.8.8.8"))
	if a != nil || b != nil {
		t.Fatalf("not deleted a=%+v b=%+v", a, b)
	}
}
