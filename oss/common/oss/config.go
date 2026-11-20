package oss

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ModePublic  = "public"
	ModePrivate = "private"

	defaultSignTTLSeconds = 600
)

// Config 是调用方注入的连接参数，不是本包去拉的配置。
// JSON/YAML 标签方便调用方从 etcd、文件或环境变量反序列化后传入 New。
type Config struct {
	// Driver 存储驱动: s3 / minio（默认）。后续可接 cos、oss、obs 等。
	Driver    string `json:"driver" yaml:"driver" mapstructure:"driver"`
	Endpoint  string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
	FileURL   string `json:"file_url" yaml:"file_url" mapstructure:"file_url"`
	AccessKey string `json:"access_key" yaml:"access_key" mapstructure:"access_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key" mapstructure:"secret_key"`
	PathStyle bool   `json:"path_style" yaml:"path_style" mapstructure:"path_style"`
	Region    string `json:"region" yaml:"region" mapstructure:"region"`
	// Buckets 桶名 → 模式。public 拼公开链；private 预签名 GET（sign_ttl 秒，默认 600）。
	// YAML 也兼容旧写法 buckets: [name]（视为 public）。
	Buckets map[string]Bucket `json:"buckets" yaml:"buckets" mapstructure:"buckets"`
}

// Bucket 单个桶的读写形态。
type Bucket struct {
	Mode    string `json:"mode" yaml:"mode" mapstructure:"mode"` // public | private
	SignTTL int    `json:"sign_ttl,omitempty" yaml:"sign_ttl" mapstructure:"sign_ttl"`
}

// PublicBuckets 测试或旧代码：若干桶名，一律 public。
func PublicBuckets(names ...string) map[string]Bucket {
	m := make(map[string]Bucket, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		m[n] = Bucket{Mode: ModePublic}
	}
	return m
}

func (c *Config) applyDefaults() {
	if c == nil {
		return
	}
	if c.Driver == "" {
		c.Driver = "s3"
	}
	if c.Region == "" {
		c.Region = "default"
	}
	for name, b := range c.Buckets {
		b.Mode = strings.ToLower(strings.TrimSpace(b.Mode))
		c.Buckets[name] = b
	}
}

// Validate 检验注入配置能否建连：必填凭证、buckets、已支持的 driver，并补默认值。
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("oss: config is nil")
	}
	c.applyDefaults()
	if c.AccessKey == "" || c.SecretKey == "" {
		return fmt.Errorf("oss: accessKey/secretKey is empty")
	}
	if len(c.BucketNames()) == 0 {
		return fmt.Errorf("oss: buckets is empty")
	}
	for name, b := range c.Buckets {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("oss: buckets has an empty name")
		}
		switch b.Mode {
		case ModePublic, ModePrivate:
		case "":
			return fmt.Errorf("oss: bucket %q: mode is required (public or private)", name)
		default:
			return fmt.Errorf("oss: bucket %q: mode %q is invalid (want public or private)", name, b.Mode)
		}
		if b.SignTTL < 0 {
			return fmt.Errorf("oss: bucket %q: sign_ttl must be >= 0", name)
		}
	}
	switch strings.ToLower(c.Driver) {
	case "s3", "minio":
	default:
		return fmt.Errorf("oss: unsupported driver %q", c.Driver)
	}
	return nil
}

// BucketNames 返回已配置的桶名（排序后，便于稳定默认值）。
func (c Config) BucketNames() []string {
	if len(c.Buckets) == 0 {
		return nil
	}
	names := make([]string, 0, len(c.Buckets))
	for n := range c.Buckets {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// FirstBucket 排序后的第一个桶名；未配置时为空。
func (c Config) FirstBucket() string {
	names := c.BucketNames()
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

// BucketByMode 返回排序后第一个 mode 匹配的桶名。同 mode 多桶时取名字更小的。
func (c Config) BucketByMode(mode string) string {
	want := strings.ToLower(strings.TrimSpace(mode))
	if want == "" {
		return ""
	}
	for _, name := range c.BucketNames() {
		if strings.EqualFold(c.Buckets[name].Mode, want) {
			return name
		}
	}
	return ""
}

// AllowBucket 检查目标是否在运维下发的 buckets 中。
func (c *Config) AllowBucket(bucket string) bool {
	if c == nil || bucket == "" || len(c.Buckets) == 0 {
		return false
	}
	_, ok := c.Buckets[bucket]
	return ok
}

// IsPrivate 该桶是否走预签名 GET。未配置或 public 为 false。
func (c Config) IsPrivate(bucket string) bool {
	b, ok := c.Buckets[bucket]
	return ok && b.Mode == ModePrivate
}

// AccessTTL 私有桶预签名 GET 有效期。public 或未知桶返回 0（调用方应拼公开链）。
func (c Config) AccessTTL(bucket string) time.Duration {
	if !c.IsPrivate(bucket) {
		return 0
	}
	sec := c.Buckets[bucket].SignTTL
	if sec <= 0 {
		sec = defaultSignTTLSeconds
	}
	return time.Duration(sec) * time.Second
}

func (c *Config) publicBase() string {
	if c == nil {
		return ""
	}
	base := strings.TrimRight(c.FileURL, "/")
	if base == "" {
		base = strings.TrimRight(c.Endpoint, "/")
	}
	return base
}

// PublicURL 按 path_style 拼公开读地址。private 桶不要走这条。
// path_style=true（MinIO）：{file_url}/{bucket}/{key}
// path_style=false（CDN / 虚拟主机）：{file_url}/{key}
func (c Config) PublicURL(bucket, key string) string {
	return JoinPublicURL(c.publicBase(), bucket, key, c.PathStyle)
}

// JoinPublicURL 公开链拼接。业务列表缓存了 file_url / path_style / bucket 后用同一规则。
func JoinPublicURL(base, bucket, key string, pathStyle bool) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	bucket = strings.TrimSpace(bucket)
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if base == "" || key == "" {
		return ""
	}
	if !pathStyle {
		return base + "/" + key
	}
	if bucket == "" {
		return ""
	}
	return base + "/" + bucket + "/" + key
}

// NormalizeBuckets 把 viper/yaml 里的 buckets 收成 map[name]Bucket。
// 支持：["name"]（视为 public）、{name: public}、{name: {mode, sign_ttl}}。
func NormalizeBuckets(raw interface{}) (map[string]Bucket, error) {
	if raw == nil {
		return nil, nil
	}
	if m, ok := raw.(map[string]Bucket); ok {
		return cloneBuckets(m), nil
	}
	out := make(map[string]Bucket)
	switch v := raw.(type) {
	case []string:
		for _, name := range v {
			if err := putBucket(out, name, Bucket{Mode: ModePublic}); err != nil {
				return nil, err
			}
		}
	case []interface{}:
		for _, item := range v {
			switch t := item.(type) {
			case string:
				if err := putBucket(out, t, Bucket{Mode: ModePublic}); err != nil {
					return nil, err
				}
			case map[string]interface{}:
				name, _ := t["name"].(string)
				b, err := bucketFromMap(t)
				if err != nil {
					return nil, err
				}
				if err := putBucket(out, name, b); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("oss: buckets list item must be a name or {name, mode}")
			}
		}
	case map[string]interface{}:
		for name, item := range v {
			b, err := bucketFromValue(item)
			if err != nil {
				return nil, fmt.Errorf("oss: bucket %q: %w", name, err)
			}
			if err := putBucket(out, name, b); err != nil {
				return nil, err
			}
		}
	case map[string]string:
		for name, mode := range v {
			if err := putBucket(out, name, Bucket{Mode: mode}); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("oss: buckets must be a list or a map")
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func cloneBuckets(src map[string]Bucket) map[string]Bucket {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]Bucket, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func putBucket(out map[string]Bucket, name string, b Bucket) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("oss: buckets has an empty name")
	}
	if _, dup := out[name]; dup {
		return fmt.Errorf("oss: duplicate bucket %q", name)
	}
	b.Mode = strings.ToLower(strings.TrimSpace(b.Mode))
	out[name] = b
	return nil
}

func bucketFromValue(item interface{}) (Bucket, error) {
	switch t := item.(type) {
	case string:
		return Bucket{Mode: t}, nil
	case map[string]interface{}:
		return bucketFromMap(t)
	case Bucket:
		return t, nil
	default:
		return Bucket{}, fmt.Errorf("want mode string or {mode, sign_ttl}")
	}
}

func bucketFromMap(m map[string]interface{}) (Bucket, error) {
	var b Bucket
	if mode, ok := m["mode"].(string); ok {
		b.Mode = mode
	}
	if ttl, ok := intFrom(m["sign_ttl"]); ok {
		b.SignTTL = ttl
	}
	return b, nil
}

func intFrom(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case uint64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

// NewStore 按 driver 创建实现。未知驱动返回错误，避免静默落到错误厂商。
func NewStore(cfg Config) (Store, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	switch strings.ToLower(cfg.Driver) {
	case "s3", "minio":
		return NewS3(cfg)
	default:
		return nil, fmt.Errorf("oss: unsupported driver %q", cfg.Driver)
	}
}
