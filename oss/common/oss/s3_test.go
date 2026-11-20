package oss

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testS3Config(endpoint string, buckets ...string) Config {
	if endpoint == "" {
		endpoint = "http://127.0.0.1:9200"
	}
	if len(buckets) == 0 {
		buckets = []string{"images"}
	}
	return Config{
		Driver:    "s3",
		Endpoint:  endpoint,
		AccessKey: "test",
		SecretKey: "test",
		PathStyle: true,
		Region:    "default",
		Buckets:   PublicBuckets(buckets...),
	}
}

func newS3Store(t *testing.T, cfg Config) *S3Store {
	t.Helper()
	store, err := NewS3(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestPresignPutSignsContentLength(t *testing.T) {
	store := newS3Store(t, testS3Config(""))

	got, err := store.PresignPut(context.Background(), PresignPutInput{
		Bucket:        "images",
		Key:           "a/b.mp4",
		ContentType:   "video/mp4",
		ContentLength: 100,
		TTL:           time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Headers["Content-Type"] != "video/mp4" {
		t.Fatalf("Content-Type: %q", got.Headers["Content-Type"])
	}
	if got.Headers["Content-Length"] != "100" {
		t.Fatalf("Content-Length: %q", got.Headers["Content-Length"])
	}
	signed := strings.ToLower(got.URL)
	if !strings.Contains(signed, "content-length") {
		t.Fatalf("signed URL missing content-length: %s", got.URL)
	}

	skip, err := store.PresignPut(context.Background(), PresignPutInput{
		Bucket:      "images",
		Key:         "a/b.mp4",
		ContentType: "video/mp4",
		TTL:         time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := skip.Headers["Content-Length"]; ok {
		t.Fatal("ContentLength<=0 should not set Content-Length header")
	}
}

func TestPresignPut(t *testing.T) {
	store := newS3Store(t, testS3Config(""))
	ctx := context.Background()

	got, err := store.PresignPut(ctx, PresignPutInput{
		Bucket:        "images",
		Key:           "/20260826/abcd.mp4",
		ContentType:   "video/mp4",
		ContentLength: 524288000,
		TTL:           10 * time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != "PUT" {
		t.Fatalf("method: %q", got.Method)
	}
	if got.URL == "" {
		t.Fatal("empty put URL")
	}
	if !strings.Contains(got.URL, "images/20260826/abcd.mp4") {
		t.Fatalf("URL missing key: %s", got.URL)
	}
	if !strings.Contains(got.URL, "X-Amz-Signature=") {
		t.Fatalf("URL missing signature: %s", got.URL)
	}
	if !strings.Contains(got.URL, "X-Amz-Expires=") {
		t.Fatalf("URL missing expires: %s", got.URL)
	}
	if got.Headers["Content-Type"] != "video/mp4" || got.Headers["Content-Length"] != "524288000" {
		t.Fatalf("headers: %+v", got.Headers)
	}

	emptyType, err := store.PresignPut(ctx, PresignPutInput{
		Bucket: "images",
		Key:    "a.bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if emptyType.Headers["Content-Type"] != "application/octet-stream" {
		t.Fatalf("default Content-Type: %q", emptyType.Headers["Content-Type"])
	}

	_, err = store.PresignPut(ctx, PresignPutInput{Key: "a.mp4"})
	if err == nil {
		t.Fatal("empty bucket should fail")
	}
	_, err = store.PresignPut(ctx, PresignPutInput{Bucket: "images"})
	if err == nil {
		t.Fatal("empty key should fail")
	}
	_, err = store.PresignPut(ctx, PresignPutInput{Bucket: "other", Key: "a.mp4"})
	if err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("foreign bucket: %v", err)
	}
}

func TestClientPresignPut(t *testing.T) {
	cli, err := New(testS3Config(""))
	if err != nil {
		t.Fatal(err)
	}
	got, err := cli.PresignPut(context.Background(), PresignPutInput{
		Bucket:        "images",
		Key:           "k.mp4",
		ContentType:   "video/mp4",
		ContentLength: 12,
		TTL:           time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != "PUT" || !strings.Contains(got.URL, "X-Amz-Signature=") {
		t.Fatalf("client presign: %+v", got)
	}
}

func TestObjectURLFollowsBucketMode(t *testing.T) {
	store := newS3Store(t, Config{
		Driver:    "s3",
		Endpoint:  "http://127.0.0.1:9200",
		FileURL:   "https://cdn.example.com",
		AccessKey: "test",
		SecretKey: "test",
		PathStyle: true,
		Buckets: map[string]Bucket{
			"images":  {Mode: ModePublic},
			"secrets": {Mode: ModePrivate, SignTTL: 120},
		},
	})
	ctx := context.Background()

	public, err := store.ObjectURL(ctx, "images", "20260826/a.png", 0)
	if err != nil {
		t.Fatal(err)
	}
	if public != "https://cdn.example.com/images/20260826/a.png" {
		t.Fatalf("path_style public: %s", public)
	}
	virtual, err := newS3Store(t, Config{
		Driver: "s3", Endpoint: "http://127.0.0.1:9200", FileURL: "https://cdn.example.com",
		AccessKey: "test", SecretKey: "test", PathStyle: false,
		Buckets: map[string]Bucket{"images": {Mode: ModePublic}},
	}).ObjectURL(ctx, "images", "20260826/a.png", 0)
	if err != nil || virtual != "https://cdn.example.com/20260826/a.png" {
		t.Fatalf("virtual-host public: %v %s", err, virtual)
	}
	ignoredTTL, err := store.ObjectURL(ctx, "images", "20260826/a.png", time.Minute)
	if err != nil || ignoredTTL != public {
		t.Fatalf("public ignores ttl: %v %s", err, ignoredTTL)
	}

	signed, err := store.ObjectURL(ctx, "secrets", "20260826/a.png", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(signed, "X-Amz-Signature=") {
		t.Fatalf("private auto-sign: %s", signed)
	}

	cli, err := New(testS3Config(""))
	if err != nil {
		t.Fatal(err)
	}
	via, err := cli.URL(ctx, "images", "k.png")
	if err != nil || strings.Contains(via, "X-Amz-Signature=") {
		t.Fatalf("client public: %v %s", err, via)
	}
	tmp, err := cli.URL(ctx, "images", "k.png", 2*time.Minute)
	if err != nil || strings.Contains(tmp, "X-Amz-Signature=") {
		t.Fatalf("client public still ignores ttl: %v %s", err, tmp)
	}
}

func TestHeadObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("method %s", r.Method)
		}
		if r.URL.Path != "/images/20260826/abcd.mp4" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	store := newS3Store(t, testS3Config(srv.URL))
	ctx := context.Background()

	head, err := store.HeadObject(ctx, "images", "20260826/abcd.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if head.Size != 100 || head.ContentType != "video/mp4" {
		t.Fatalf("head: %+v", head)
	}

	_, err = store.HeadObject(ctx, "images", "missing.mp4")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}

	_, err = store.HeadObject(ctx, "", "k")
	if err == nil {
		t.Fatal("empty bucket should fail")
	}
	_, err = store.HeadObject(ctx, "images", "")
	if err == nil {
		t.Fatal("empty key should fail")
	}

	cli, err := New(testS3Config(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	viaClient, err := cli.Head(ctx, "images", "20260826/abcd.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if viaClient.Size != 100 {
		t.Fatalf("client head: %+v", viaClient)
	}
}

func TestDeleteObject(t *testing.T) {
	var deleted string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method %s", r.Method)
		}
		deleted = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	store := newS3Store(t, testS3Config(srv.URL))
	ctx := context.Background()

	if err := store.DeleteObject(ctx, "images", "/20260826/abcd.mp4"); err != nil {
		t.Fatal(err)
	}
	if deleted != "/images/20260826/abcd.mp4" {
		t.Fatalf("deleted path: %q", deleted)
	}

	if err := store.DeleteObject(ctx, "", "k"); err == nil {
		t.Fatal("empty bucket should fail")
	}
	if err := store.DeleteObject(ctx, "images", ""); err == nil {
		t.Fatal("empty key should fail")
	}
	if err := store.DeleteObject(ctx, "other", "k"); err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("foreign bucket: %v", err)
	}

	cli, err := New(testS3Config(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Delete(ctx, "images", "20260826/abcd.mp4"); err != nil {
		t.Fatal(err)
	}
}

type coreOnlyStore struct{}

func (coreOnlyStore) PutObject(context.Context, PutObjectInput) error { return nil }
func (coreOnlyStore) ObjectExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (coreOnlyStore) ObjectURL(context.Context, string, string, time.Duration) (string, error) {
	return "", nil
}

func TestPresignStoreOptional(t *testing.T) {
	c := &Client{impl: coreOnlyStore{}}
	_, err := c.PresignPut(context.Background(), PresignPutInput{Bucket: "b", Key: "k"})
	if !errors.Is(err, ErrPresignUnsupported) {
		t.Fatalf("PresignPut: %v", err)
	}
	_, err = c.Head(context.Background(), "b", "k")
	if !errors.Is(err, ErrPresignUnsupported) {
		t.Fatalf("Head: %v", err)
	}
	if err := c.Delete(context.Background(), "b", "k"); !errors.Is(err, ErrPresignUnsupported) {
		t.Fatalf("Delete: %v", err)
	}
}
