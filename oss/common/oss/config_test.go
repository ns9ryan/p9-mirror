package oss

import (
	"testing"
	"time"
)

func TestAllowBucketRequiresConfiguredList(t *testing.T) {
	var nilCfg *Config
	if nilCfg.AllowBucket("images") {
		t.Fatal("nil config should deny")
	}
	empty := Config{}
	if empty.AllowBucket("images") {
		t.Fatal("empty buckets should deny")
	}
	cfg := Config{Buckets: PublicBuckets("images")}
	if !cfg.AllowBucket("images") {
		t.Fatal("listed bucket should allow")
	}
	if cfg.AllowBucket("other") || cfg.AllowBucket("") {
		t.Fatal("unlisted or empty bucket should deny")
	}
}

func TestValidateRequiresBuckets(t *testing.T) {
	cfg := Config{AccessKey: "a", SecretKey: "b"}
	if err := cfg.Validate(); err == nil || err.Error() != "oss: buckets is empty" {
		t.Fatalf("empty buckets: %v", err)
	}
	cfg.Buckets = map[string]Bucket{"  ": {Mode: ModePublic}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("whitespace-only buckets should fail")
	}
	cfg.Buckets = PublicBuckets("images")
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRequiresBucketMode(t *testing.T) {
	cfg := Config{AccessKey: "a", SecretKey: "b", Buckets: map[string]Bucket{
		"images": {},
	}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("missing mode should fail")
	}
	cfg.Buckets["images"] = Bucket{Mode: "signed"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("unknown mode should fail")
	}
	cfg.Buckets["images"] = Bucket{Mode: ModePrivate, SignTTL: 30}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if !cfg.IsPrivate("images") || cfg.AccessTTL("images") != 30*time.Second {
		t.Fatalf("private ttl: private=%v ttl=%s", cfg.IsPrivate("images"), cfg.AccessTTL("images"))
	}
}

func TestNormalizeBuckets(t *testing.T) {
	fromList, err := NormalizeBuckets([]string{"images", "assets"})
	if err != nil || fromList["images"].Mode != ModePublic || fromList["assets"].Mode != ModePublic {
		t.Fatalf("list: %+v %v", fromList, err)
	}
	fromShort, err := NormalizeBuckets(map[string]string{
		"images":  "public",
		"secrets": "private",
	})
	if err != nil || fromShort["secrets"].Mode != ModePrivate {
		t.Fatalf("shorthand: %+v %v", fromShort, err)
	}
	fromObj, err := NormalizeBuckets(map[string]interface{}{
		"secrets": map[string]interface{}{"mode": "private", "sign_ttl": 120},
	})
	if err != nil || fromObj["secrets"].Mode != ModePrivate || fromObj["secrets"].SignTTL != 120 {
		t.Fatalf("object: %+v %v", fromObj, err)
	}
}

func TestAccessTTLPublicIsZero(t *testing.T) {
	cfg := Config{Buckets: map[string]Bucket{
		"images":  {Mode: ModePublic},
		"secrets": {Mode: ModePrivate},
	}}
	if cfg.AccessTTL("images") != 0 {
		t.Fatal("public ttl should be 0")
	}
	if cfg.AccessTTL("secrets") != 600*time.Second {
		t.Fatalf("default private ttl: %s", cfg.AccessTTL("secrets"))
	}
}

func TestJoinPublicURLFollowsPathStyle(t *testing.T) {
	base, bucket, key := "https://cdn.example.com", "oss-bucket", "public/a.png"
	if got := JoinPublicURL(base, bucket, key, true); got != "https://cdn.example.com/oss-bucket/public/a.png" {
		t.Fatalf("path style: %s", got)
	}
	if got := JoinPublicURL(base, bucket, key, false); got != "https://cdn.example.com/public/a.png" {
		t.Fatalf("virtual host: %s", got)
	}
	if JoinPublicURL(base, "", key, true) != "" {
		t.Fatal("path style needs bucket")
	}
	cfg := Config{FileURL: base, PathStyle: false}
	if got := cfg.PublicURL(bucket, key); got != "https://cdn.example.com/public/a.png" {
		t.Fatalf("config: %s", got)
	}
}

func TestBucketByMode(t *testing.T) {
	cfg := Config{Buckets: map[string]Bucket{
		"z-pub": {Mode: ModePublic},
		"a-pri": {Mode: ModePrivate},
		"b-pri": {Mode: ModePrivate},
	}}
	if got := cfg.BucketByMode(ModePrivate); got != "a-pri" {
		t.Fatalf("private: %s", got)
	}
	if cfg.BucketByMode(ModePublic) != "z-pub" {
		t.Fatal(cfg.BucketByMode(ModePublic))
	}
	if cfg.BucketByMode("missing") != "" {
		t.Fatal("unknown mode")
	}
}
