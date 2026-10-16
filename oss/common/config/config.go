package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"oa.98ent.com/p9/common/i18n"
	"oa.98ent.com/p9/oss/common/oss"
)

var (
	ErrInvalidScene = errors.New("invalid scene")
	sceneNameRe     = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
)

const (
	defaultMaxUploadSizeMB   = 32
	defaultMultipartMemoryMB = 32
	defaultPresignMaxSizeMB  = 2048
	defaultPresignTTLSeconds = 600
	// MaxRequestOverhead multipart 字段等额外开销，叠在文件上限上给 MaxBytesReader。
	MaxRequestOverhead int64 = 1 << 20
)

// AppConfig 是本地 YAML 或 etcd `{KEY}/service` 的整份配置：HTTP、上传业务、存储连接、可选 core / 本服务 RPC。
type AppConfig struct {
	Port     string         `mapstructure:"port"`
	I18n     i18n.Config    `mapstructure:"i18n"`
	CROSConf CORSConfig     `mapstructure:"cros_conf"`
	Upload   UploadConfig   `mapstructure:"upload"`
	OSS      oss.Config     `mapstructure:"oss"`
	Database DatabaseConfig `mapstructure:"database"`
	CoreRPC  CoreRPCConfig  `mapstructure:"core_rpc"`
	RPC      RPCConfig      `mapstructure:"rpc"`
	OssRPC   OssRPCConfig   `mapstructure:"oss_rpc"`
}

// CORSConfig 对应 yaml cros_conf.address，给 rest.WithCors。旧 key CROSConf 仍能解析。
type CORSConfig struct {
	Address string `mapstructure:"address"`
}

// DatabaseConfig 可选。配了才落文件档案、才能用服务商管理。
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// UploadConfig 上传业务规则。单位用 MB，方便在 yaml 里改。
type UploadConfig struct {
	MaxSizeMB         int                  `mapstructure:"max_size_mb"`
	MultipartMemoryMB int                  `mapstructure:"multipart_memory_mb"`
	AllowedExtensions []string             `mapstructure:"allowed_extensions"` // 空或不配：允许全部后缀
	PresignMaxSizeMB  int                  `mapstructure:"presign_max_size_mb"`
	PresignTTL        int                  `mapstructure:"presign_ttl"`
	DefaultScene      string               `mapstructure:"default_scene"`
	Scenes            map[string]SceneYAML `mapstructure:"scenes"` // scene → visibility；启动时按桶 mode 对齐
}

// SceneYAML 进程级业务目录：只声明公开/私有，不写桶名。
type SceneYAML struct {
	Visibility string `mapstructure:"visibility"`
}

// Settings 启动后使用的运行时配置（已填默认值，大小已换成字节）。
type Settings struct {
	Port        string
	Upload      UploadSettings
	OSS         oss.Config
	DatabaseDSN string
}

// UploadSettings HTTP 层使用的上传限制（由 UploadConfig 换算，大小为字节）。
type UploadSettings struct {
	MaxSize           int64         // POST /file/upload 中转上限
	MultipartMemory   int64         // multipart 内存阈值，超出落临时文件
	AllowedExtensions []string      // 空=全部允许；配了则只允许列表中的后缀
	PresignMaxSize    int64         // 预签名直传上限（/upload/presign、complete）
	PresignTTL        time.Duration // 预签名 PUT 有效期（putUrl 过期时间）
	DefaultScene      string        // 请求未带 scene 时使用
	Scenes            map[string]SceneSettings
	sceneErr          error
}

// SceneSettings 运行时的 scene：visibility 来自 yaml，Bucket 按 oss.buckets.mode 对齐。
type SceneSettings struct {
	Visibility string
	Bucket     string
}

// applyDefaults 补 OSS driver/region 以及 upload 数值默认值。
func (d *AppConfig) applyDefaults() {
	if d == nil {
		return
	}
	d.Upload.applyDefaults()
	if d.OSS.Driver == "" {
		d.OSS.Driver = "s3"
	}
	if d.OSS.Region == "" {
		d.OSS.Region = "default"
	}
}

// applyDefaults 补上传大小与 presign TTL 的缺省值（MB / 秒）。
func (u *UploadConfig) applyDefaults() {
	if u == nil {
		return
	}
	if u.MaxSizeMB <= 0 {
		u.MaxSizeMB = defaultMaxUploadSizeMB
	}
	if u.MultipartMemoryMB <= 0 {
		u.MultipartMemoryMB = defaultMultipartMemoryMB
	}
	if u.PresignMaxSizeMB <= 0 {
		u.PresignMaxSizeMB = defaultPresignMaxSizeMB
	}
	if u.PresignTTL <= 0 {
		u.PresignTTL = defaultPresignTTLSeconds
	}
}

// Settings 换成运行时配置：大小变字节，scenes 按 visibility 对齐到桶。
func (d *AppConfig) Settings() Settings {
	d.applyDefaults()
	s := Settings{
		Port:        d.Port,
		Upload:      d.Upload.settings(),
		OSS:         d.OSS,
		DatabaseDSN: strings.TrimSpace(d.Database.DSN),
	}
	s.applyEnv()
	s.applySceneDefaults()
	return s
}

func (s *Settings) applyEnv() {
	if env := strings.TrimSpace(os.Getenv("POSTGRES_DSN")); env != "" {
		s.DatabaseDSN = env
	}
	if env := strings.TrimSpace(os.Getenv("OSS_ENDPOINT")); env != "" {
		s.OSS.Endpoint = env
	}
	if env := strings.TrimSpace(os.Getenv("OSS_FILE_URL")); env != "" {
		s.OSS.FileURL = env
	}
}

// applySceneDefaults 未配 default_scene 时用 public；未配 scenes 时挂 public。再按桶 mode 对齐。
func (s *Settings) applySceneDefaults() {
	if s.Upload.DefaultScene == "" {
		s.Upload.DefaultScene = "public"
	}
	if s.Upload.sceneErr == nil && len(s.Upload.Scenes) == 0 {
		s.Upload.Scenes = map[string]SceneSettings{
			"public": {Visibility: oss.ModePublic},
		}
	}
	s.bindScenes()
}

// bindScenes 把每个 scene 的 visibility 对齐到排序后第一个同 mode 的桶。
func (s *Settings) bindScenes() {
	if s == nil || len(s.Upload.Scenes) == 0 {
		return
	}
	next := make(map[string]SceneSettings, len(s.Upload.Scenes))
	for name, sc := range s.Upload.Scenes {
		vis := strings.ToLower(strings.TrimSpace(sc.Visibility))
		if vis == "" {
			vis = oss.ModePublic
		}
		sc.Visibility = vis
		sc.Bucket = s.OSS.BucketByMode(vis)
		next[name] = sc
	}
	s.Upload.Scenes = next
}

// ValidateUpload 只校验 scene 目录，不要求 oss.buckets。
func (s Settings) ValidateUpload() error {
	if s.Upload.sceneErr != nil {
		return s.Upload.sceneErr
	}
	if len(s.Upload.Scenes) == 0 {
		return errors.New("upload.scenes is empty")
	}
	if !validSceneName(s.Upload.DefaultScene) {
		return fmt.Errorf("upload.default_scene %q is invalid", s.Upload.DefaultScene)
	}
	if _, ok := s.Upload.Scenes[s.Upload.DefaultScene]; !ok {
		return fmt.Errorf("upload.default_scene %q is not listed in upload.scenes", s.Upload.DefaultScene)
	}
	for name, p := range s.Upload.Scenes {
		if !validSceneName(name) {
			return fmt.Errorf("scene %q is invalid (use [a-z][a-z0-9_]*, not a bucket name)", name)
		}
		vis := strings.ToLower(strings.TrimSpace(p.Visibility))
		if vis != oss.ModePublic && vis != oss.ModePrivate {
			return fmt.Errorf("upload.scenes[%s].visibility is required (public or private)", name)
		}
	}
	return nil
}

// Validate 在 ValidateUpload 之上再校验桶 mode，以及每个 scene 都能对齐到同 mode 的桶。
func (s Settings) Validate() error {
	if err := s.ValidateUpload(); err != nil {
		return err
	}
	if len(s.OSS.BucketNames()) == 0 {
		return errors.New("buckets is empty")
	}
	allowed := make(map[string]struct{}, len(s.OSS.Buckets))
	for _, b := range s.OSS.BucketNames() {
		allowed[b] = struct{}{}
	}
	for name, p := range s.Upload.Scenes {
		if err := validateSceneBinding(name, p, allowed); err != nil {
			return err
		}
	}
	for name, b := range s.OSS.Buckets {
		mode := strings.ToLower(strings.TrimSpace(b.Mode))
		if mode != oss.ModePublic && mode != oss.ModePrivate {
			return fmt.Errorf("oss.buckets[%s].mode is required (public or private)", name)
		}
	}
	return nil
}

func validateSceneBinding(name string, p SceneSettings, buckets map[string]struct{}) error {
	if !validSceneName(name) {
		return fmt.Errorf("scene %q is invalid (use [a-z][a-z0-9_]*, not a bucket name)", name)
	}
	if _, isBucket := buckets[name]; isBucket {
		return fmt.Errorf("scene %q must not be a bucket name", name)
	}
	vis := strings.ToLower(strings.TrimSpace(p.Visibility))
	if vis != oss.ModePublic && vis != oss.ModePrivate {
		return fmt.Errorf("upload.scenes[%s].visibility is required (public or private)", name)
	}
	if p.Bucket == "" {
		return fmt.Errorf("scene %q needs a bucket with mode %s", name, vis)
	}
	if _, ok := buckets[p.Bucket]; !ok {
		return fmt.Errorf("scene %q bucket %q is not in oss.buckets", name, p.Bucket)
	}
	return nil
}

// validSceneName scene 须为 [a-z][a-z0-9_]*，最长 32。
func validSceneName(s string) bool {
	return sceneNameRe.MatchString(s)
}

// ResolveScene 把请求里的 scene 映射到桶。空则用 DefaultScene。未知值（含桶名）返回 ErrInvalidScene。
func (u UploadSettings) ResolveScene(raw string) (scene, bucket string, err error) {
	p := strings.ToLower(strings.TrimSpace(raw))
	if p == "" {
		p = u.DefaultScene
	}
	if !validSceneName(p) {
		return "", "", ErrInvalidScene
	}
	cfg, ok := u.Scenes[p]
	if !ok {
		return "", "", ErrInvalidScene
	}
	if cfg.Bucket == "" {
		return "", "", fmt.Errorf("scene %q has empty bucket", p)
	}
	return p, cfg.Bucket, nil
}

// settings 把 YAML 里的 MB/秒换成字节和 Duration，并解析 scene 目录。
func (u UploadConfig) settings() UploadSettings {
	exts := make([]string, 0, len(u.AllowedExtensions))
	seen := make(map[string]struct{}, len(u.AllowedExtensions))
	for _, raw := range u.AllowedExtensions {
		ext := normalizeExt(raw)
		if ext == "" {
			continue
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		exts = append(exts, ext)
	}
	scenes, parseErr := parseSceneCatalog(u.Scenes)
	return UploadSettings{
		MaxSize:           int64(u.MaxSizeMB) << 20,
		MultipartMemory:   int64(u.MultipartMemoryMB) << 20,
		AllowedExtensions: exts,
		PresignMaxSize:    int64(u.PresignMaxSizeMB) << 20,
		PresignTTL:        time.Duration(u.PresignTTL) * time.Second,
		DefaultScene:      strings.ToLower(strings.TrimSpace(u.DefaultScene)),
		Scenes:            scenes,
		sceneErr:          parseErr,
	}
}

// parseSceneCatalog 读进程 yaml：scene → { visibility }。
func parseSceneCatalog(raw map[string]SceneYAML) (map[string]SceneSettings, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make(map[string]SceneSettings, len(raw))
	for rawName, spec := range raw {
		name := strings.ToLower(strings.TrimSpace(rawName))
		if name == "" {
			return nil, errors.New("upload.scenes has an empty scene name")
		}
		vis := strings.ToLower(strings.TrimSpace(spec.Visibility))
		if vis != oss.ModePublic && vis != oss.ModePrivate {
			return nil, fmt.Errorf("upload.scenes[%s].visibility is required (public or private)", name)
		}
		out[name] = SceneSettings{Visibility: vis}
	}
	return out, nil
}

// normalizeExt 统一成小写且带点，如 jpg → .jpg。
func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}
