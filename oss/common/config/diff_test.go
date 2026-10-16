package config

import (
	"strings"
	"testing"
	"time"

	"oa.98ent.com/p9/oss/common/oss"
)

func TestDiffSettings(t *testing.T) {
	prev := Settings{
		Port: "8020",
		Upload: UploadSettings{
			MaxSize:           32 << 20,
			MultipartMemory:   32 << 20,
			PresignMaxSize:    2048 << 20,
			PresignTTL:        600 * time.Second,
			AllowedExtensions: []string{".jpg"},
		},
		OSS: oss.Config{
			Driver:    "s3",
			Endpoint:  "http://127.0.0.1:9200",
			FileURL:   "http://cdn",
			AccessKey: "ak",
			SecretKey: "sk",
			PathStyle: true,
			Buckets:   oss.PublicBuckets("images"),
		},
	}
	next := prev
	next.Port = "80"
	next.Upload.MaxSize = 64 << 20
	next.Upload.AllowedExtensions = nil
	next.OSS.SecretKey = "sk2"
	next.OSS.Buckets = oss.PublicBuckets("images", "assets")
	next.Upload.DefaultScene = "public"
	next.Upload.Scenes = map[string]SceneSettings{
		"public": {Visibility: oss.ModePublic, Bucket: "images"},
		"kyc":    {Visibility: oss.ModePrivate, Bucket: "assets"},
	}

	got := strings.Join(DiffSettings(prev, next), "\n")
	for _, want := range []string{
		`port: 8020 -> 80 （监听端口需重启进程才生效）`,
		`upload.max_size_mb: 32 -> 64`,
		`upload.allowed_extensions: .jpg -> (空=全部允许)`,
		`oss.secret_key: 已变更（不打印明文）`,
		`oss.buckets: images:public -> assets:public,images:public`,
		`upload.scenes: (空) -> kyc:private@assets,public:public@images`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "sk2") || strings.Contains(got, "oss.endpoint") {
		t.Fatalf("should not dump secret or unchanged fields:\n%s", got)
	}
	if DiffSettings(prev, prev) != nil {
		t.Fatal("identical config should have no diff")
	}
}
