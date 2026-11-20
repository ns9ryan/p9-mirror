package config

import (
	"bytes"
	"testing"
	"time"

	"oa.98ent.com/p9/oss/common/oss"

	"github.com/spf13/viper"
)

func unmarshalYAML(t *testing.T, raw string) AppConfig {
	t.Helper()
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(raw)); err != nil {
		t.Fatal(err)
	}
	applyLegacyConfigAliases(v)
	var got AppConfig
	if err := v.Unmarshal(&got); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestUnmarshalSnakeCase(t *testing.T) {
	got := unmarshalYAML(t, `
port: "8020"
upload:
  max_size_mb: 16
  multipart_memory_mb: 8
  presign_max_size_mb: 100
  presign_ttl: 120
  default_scene: public
  scenes:
    public:
      visibility: public
    kyc:
      visibility: private
  allowed_extensions:
    - jpg
oss:
  driver: s3
  endpoint: "http://127.0.0.1:9200"
  file_url: "http://cdn.example"
  access_key: ak
  secret_key: sk
  path_style: true
  buckets:
    - images
    - oss-bucket-private
`)
	if got.Port != "8020" || got.Upload.MaxSizeMB != 16 || got.Upload.PresignTTL != 120 {
		t.Fatalf("upload: %+v", got)
	}
	if got.Upload.DefaultScene != "public" || got.Upload.Scenes["kyc"].Visibility != "private" {
		t.Fatalf("scenes: %+v", got.Upload)
	}
	if got.OSS.FileURL != "http://cdn.example" || !got.OSS.PathStyle {
		t.Fatalf("oss: %+v", got.OSS)
	}
	if got.OSS.AccessKey != "ak" || len(got.OSS.Buckets) != 2 {
		t.Fatalf("oss keys: %+v", got.OSS)
	}
	if got.OSS.Buckets["images"].Mode != oss.ModePublic {
		t.Fatalf("legacy list should coerce to public: %+v", got.OSS.Buckets)
	}
}

func TestUnmarshalCROSConf(t *testing.T) {
	got := unmarshalYAML(t, `
cros_conf:
  address: 'https://admin.example.com'
`)
	if got.CROSConf.Address != "https://admin.example.com" {
		t.Fatalf("cros: %+v", got.CROSConf)
	}
}

func TestUnmarshalLegacyCROSConf(t *testing.T) {
	got := unmarshalYAML(t, `
CROSConf:
  Address: 'https://admin.example.com'
`)
	if got.CROSConf.Address != "https://admin.example.com" {
		t.Fatalf("legacy cros: %+v", got.CROSConf)
	}
}

func TestUnmarshalLegacyCamelCase(t *testing.T) {
	got := unmarshalYAML(t, `
upload:
  maxSizeMB: 24
oss:
  fileUrl: "http://old"
  accessKey: ak
  secretKey: sk
  pathstyle: true
  buckets:
    - images
`)
	if got.Upload.MaxSizeMB != 24 {
		t.Fatalf("maxSizeMB alias: %d", got.Upload.MaxSizeMB)
	}
	if got.OSS.FileURL != "http://old" || !got.OSS.PathStyle {
		t.Fatalf("legacy oss: %+v", got.OSS)
	}
}

func TestSceneDefaultsAndResolve(t *testing.T) {
	cfg := unmarshalYAML(t, `
oss:
  buckets:
    - images
`)
	s := cfg.Settings()
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if s.Upload.DefaultScene != "public" || s.Upload.Scenes["public"].Bucket != "images" {
		t.Fatalf("auto public: %+v", s.Upload)
	}
	name, bucket, err := s.Upload.ResolveScene("")
	if err != nil || name != "public" || bucket != "images" {
		t.Fatalf("empty scene: %s %s %v", name, bucket, err)
	}
	if _, _, err := s.Upload.ResolveScene("kyc"); err == nil {
		t.Fatal("unknown scene should fail")
	}
	if _, _, err := s.Upload.ResolveScene("images"); err == nil {
		t.Fatal("bucket name should fail")
	}
}

func TestSceneMap(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  default_scene: public
  scenes:
    public:
      visibility: public
    avatar:
      visibility: public
    kyc:
      visibility: private
oss:
  buckets:
    images:
      mode: public
    oss-bucket-private:
      mode: private
`)
	s := cfg.Settings()
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	_, bucket, err := s.Upload.ResolveScene("avatar")
	if err != nil || bucket != "images" {
		t.Fatalf("avatar: %s %v", bucket, err)
	}
	_, bucket, err = s.Upload.ResolveScene("KYC")
	if err != nil || bucket != "oss-bucket-private" {
		t.Fatalf("kyc: %s %v", bucket, err)
	}
	if _, _, err := s.Upload.ResolveScene("oss-bucket-private"); err == nil {
		t.Fatal("hyphenated bucket as scene should fail")
	}
}

func TestValidateRejectsSceneNamedAsBucket(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  scenes:
    images:
      visibility: public
oss:
  buckets:
    - images
`)
	s := cfg.Settings()
	if err := s.Validate(); err == nil {
		t.Fatal("scene key equal to bucket name should fail")
	}
}

func TestValidateRejectsMissingModeBucket(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  scenes:
    kyc:
      visibility: private
oss:
  buckets:
    images:
      mode: public
`)
	s := cfg.Settings()
	if err := s.Validate(); err == nil {
		t.Fatal("private scene without private bucket should fail")
	}
}

func TestSceneSameVisibilityShareBucket(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  scenes:
    public:
      visibility: public
    avatar:
      visibility: public
oss:
  buckets:
    images:
      mode: public
`)
	s := cfg.Settings()
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if s.Upload.Scenes["public"].Bucket != "images" || s.Upload.Scenes["avatar"].Bucket != "images" {
		t.Fatalf("share: %+v", s.Upload.Scenes)
	}
}

func TestUnmarshalSceneShorthand(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  scenes:
    public: public
    kyc: private
oss:
  buckets:
    images:
      mode: public
    secrets:
      mode: private
`)
	s := cfg.Settings()
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, bucket, err := s.Upload.ResolveScene("kyc"); err != nil || bucket != "secrets" {
		t.Fatalf("kyc: %s %v", bucket, err)
	}
}

func TestUnmarshalBucketModes(t *testing.T) {
	got := unmarshalYAML(t, `
oss:
  buckets:
    images:
      mode: public
    secrets:
      mode: private
      sign_ttl: 120
`)
	if got.OSS.Buckets["images"].Mode != oss.ModePublic {
		t.Fatalf("images: %+v", got.OSS.Buckets["images"])
	}
	if !got.OSS.IsPrivate("secrets") || got.OSS.AccessTTL("secrets") != 120*time.Second {
		t.Fatalf("secrets: %+v ttl=%s", got.OSS.Buckets["secrets"], got.OSS.AccessTTL("secrets"))
	}
	short := unmarshalYAML(t, `
oss:
  buckets:
    images: public
    secrets: private
`)
	if short.OSS.Buckets["secrets"].Mode != oss.ModePrivate {
		t.Fatalf("shorthand: %+v", short.OSS.Buckets)
	}
}

func TestValidateRequiresOSSBucketMode(t *testing.T) {
	cfg := unmarshalYAML(t, `
oss:
  buckets:
    images: {}
`)
	s := cfg.Settings()
	if err := s.Validate(); err == nil {
		t.Fatal("missing bucket mode should fail")
	}
}

func TestLoadFromFile(t *testing.T) {
	var d AppConfig
	if _, err := d.loadFromFile("rpc/etc/oss.yaml"); err != nil {
		t.Fatal(err)
	}
	if d.RPC.ListenOn != "0.0.0.0:8021" || d.Database.DSN == "" {
		t.Fatalf("rpc: listen=%q dsn=%q", d.RPC.ListenOn, d.Database.DSN)
	}
	if d.Upload.DefaultScene != "" || len(d.OSS.Buckets) != 0 {
		t.Fatalf("rpc yaml should not include upload/oss: %+v %+v", d.Upload, d.OSS)
	}
	var api AppConfig
	if _, err := api.loadFromFile("api/etc/oss.yaml"); err != nil {
		t.Fatal(err)
	}
	if api.Port != "8020" {
		t.Fatalf("api port=%q", api.Port)
	}
	as := api.Settings()
	if err := as.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, bucket, err := as.Upload.ResolveScene("public"); err != nil || bucket != "oss-bucket" {
		t.Fatalf("public: %s %v", bucket, err)
	}
	if _, _, err := as.Upload.ResolveScene("idcard"); err == nil {
		t.Fatal("unlisted scene should fail")
	}
}

func TestValidateUploadNoOSS(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  default_scene: public
  scenes:
    public:
      visibility: public
`)
	s := cfg.Settings()
	if err := s.ValidateUpload(); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(); err == nil {
		t.Fatal("empty oss should fail full Validate")
	}
}

func TestLoadInitDefaultSettings(t *testing.T) {
	spec, err := LoadInitProviderTemplate("rpc/etc/init_default_settings.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if spec.OSS.AccessKey == "" || spec.OSS.Buckets["oss-bucket"].Mode != oss.ModePublic {
		t.Fatalf("template: %+v", spec)
	}
	if _, ok := spec.OSS.Buckets["oss-bucket-private"]; ok {
		t.Fatal("default template should not include private bucket yet")
	}
}

func TestResolveConfigFileMissing(t *testing.T) {
	if _, err := resolveConfigFile("no-such-config.yaml"); err == nil {
		t.Fatal("expected missing file")
	}
}

func TestApplyProviderEmptyBaseOSS(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  default_scene: public
  scenes:
    public:
      visibility: public
`)
	next, err := ApplyProvider(cfg.Settings(), ProviderSpec{
		OSS: oss.Config{
			AccessKey: "ak",
			Buckets:   map[string]oss.Bucket{"oss-bucket": {Mode: oss.ModePublic}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if next.Upload.Scenes["public"].Bucket != "oss-bucket" {
		t.Fatalf("bind: %+v", next.Upload.Scenes)
	}
}

func TestApplyProviderOverridesOSSKeepsLimits(t *testing.T) {
	cfg := unmarshalYAML(t, `
upload:
  max_size_mb: 16
  default_scene: public
  scenes:
    public:
      visibility: public
    avatar:
      visibility: public
oss:
  access_key: ak
  secret_key: sk
  buckets:
    images:
      mode: public
`)
	base := cfg.Settings()
	if err := base.Validate(); err != nil {
		t.Fatal(err)
	}
	next, err := ApplyProvider(base, ProviderSpec{
		OSS: oss.Config{
			Driver:    "s3",
			AccessKey: "ak2",
			SecretKey: "sk2",
			PathStyle: true,
			Buckets: map[string]oss.Bucket{
				"prod": {Mode: oss.ModePublic},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if next.OSS.AccessKey != "ak2" || next.OSS.Buckets["prod"].Mode != oss.ModePublic {
		t.Fatalf("oss: %+v", next.OSS)
	}
	if next.Upload.DefaultScene != "public" || next.Upload.Scenes["avatar"].Bucket != "prod" {
		t.Fatalf("upload: %+v", next.Upload)
	}
	if next.Upload.MaxSize != base.Upload.MaxSize {
		t.Fatalf("max size should stay: %d vs %d", next.Upload.MaxSize, base.Upload.MaxSize)
	}
	if _, _, err := next.Upload.ResolveScene("public"); err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalCoreRPC(t *testing.T) {
	got := unmarshalYAML(t, `
port: "8020"
core_rpc:
  endpoints:
    - 127.0.0.1:8080
  timeout: 8000
oss:
  buckets:
    images:
      mode: public
`)
	if len(got.CoreRPC.Endpoints) != 1 || got.CoreRPC.Endpoints[0] != "127.0.0.1:8080" {
		t.Fatalf("endpoints: %+v", got.CoreRPC)
	}
	if got.CoreRPC.Timeout != 8000 {
		t.Fatalf("timeout: %d", got.CoreRPC.Timeout)
	}
}

func TestUnmarshalRPC(t *testing.T) {
	got := unmarshalYAML(t, `
port: "8020"
rpc:
  listen_on: 0.0.0.0:8021
oss:
  buckets:
    images:
      mode: public
`)
	if got.RPC.ListenOn != "0.0.0.0:8021" {
		t.Fatalf("rpc: %+v", got.RPC)
	}
}

func TestUnmarshalOssRPC(t *testing.T) {
	got := unmarshalYAML(t, `
port: "8020"
oss_rpc:
  endpoints:
    - 127.0.0.1:8021
  timeout: 300000
oss:
  buckets:
    images:
      mode: public
`)
	if len(got.OssRPC.Endpoints) != 1 || got.OssRPC.Endpoints[0] != "127.0.0.1:8021" {
		t.Fatalf("oss_rpc: %+v", got.OssRPC)
	}
	if got.OssRPC.Timeout != 300000 {
		t.Fatalf("timeout: %d", got.OssRPC.Timeout)
	}
}
