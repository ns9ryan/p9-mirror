package config

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"oa.98ent.com/p9/oss/common/oss"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const configPathSuffix = "/service"

var (
	once          sync.Once
	mu            sync.RWMutex
	inst          *AppConfig
	onChange      func(Settings)
	etcdClient    *clientv3.Client
	etcdConfigKey string
)

// OnChange 注册热更新回调。解析成功后把新 Settings 交给调用方（Handler / oss.Client.Apply）。
func OnChange(fn func(Settings)) {
	mu.Lock()
	onChange = fn
	mu.Unlock()
}

// initEtcdClient 用环境变量建立 etcd 连接。
func (d *AppConfig) initEtcdClient() error {
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(os.Getenv("ETCD_ENDPOINTS"), ","),
		DialTimeout: 8 * time.Second,
		Username:    os.Getenv("ETCD_USERNAME"),
		Password:    os.Getenv("ETCD_PASSWORD"),
	})
	if err != nil {
		return fmt.Errorf("创建 etcd 客户端失败: %v", err)
	}
	return nil
}

// decodeYAML 解析 YAML 到 AppConfig，并做旧字段别名与默认值。
func decodeYAML(raw []byte, dst *AppConfig) error {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBuffer(raw)); err != nil {
		return err
	}
	applyLegacyConfigAliases(v)
	if err := v.Unmarshal(dst); err != nil {
		return err
	}
	dst.applyDefaults()
	return nil
}

// resolveConfigFile 解析 -config 路径；相对路径还会尝试 config/ 目录。
func resolveConfigFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	candidates := []string{path}
	if !filepath.IsAbs(path) {
		dir, _ := os.Getwd()
		for i := 0; i < 6 && dir != ""; i++ {
			candidates = append(candidates,
				filepath.Join(dir, path),
				filepath.Join(dir, "etc", path),
				filepath.Join(dir, "etc", filepath.Base(path)),
				filepath.Join(dir, "config", path),
				filepath.Join(dir, "config", filepath.Base(path)),
			)
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	seen := make(map[string]struct{}, len(candidates))
	for _, p := range candidates {
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		st, err := os.Stat(p)
		if err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("配置文件不存在: %s", path)
}

// loadFromFile 从本地 YAML 加载配置，返回实际用到的路径。
func (d *AppConfig) loadFromFile(path string) (string, error) {
	resolved, err := resolveConfigFile(path)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return "", fmt.Errorf("读取配置文件 %s: %w", resolved, err)
	}
	if err := decodeYAML(raw, d); err != nil {
		return "", fmt.Errorf("解析配置文件 %s: %w", resolved, err)
	}
	return resolved, nil
}

// readConfig 从 etcd 拉取并解析当前配置。
func (d *AppConfig) readConfig() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := etcdClient.Get(ctx, etcdConfigKey)
	if err != nil {
		panic(fmt.Sprintf("从 etcd 获取配置失败: %v", err))
	}
	if len(resp.Kvs) == 0 {
		panic(fmt.Sprintf("etcd 中未找到配置:%v,", etcdConfigKey))
	}
	if err := decodeYAML(resp.Kvs[0].Value, d); err != nil {
		panic(fmt.Sprintf("解析配置失败: %v", err))
	}
}

// watchConfig 监听 etcd PUT，校验通过后触发 OnChange。
func (d *AppConfig) watchConfig() {
	watchChan := etcdClient.Watch(context.Background(), etcdConfigKey)
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			if event.Type != clientv3.EventTypePut {
				continue
			}
			newConfig := &AppConfig{}
			if err := decodeYAML(event.Kv.Value, newConfig); err != nil {
				fmt.Printf("解析新配置失败: %v\n", err)
				continue
			}
			mu.Lock()
			inst = newConfig
			cb := onChange
			mu.Unlock()
			if cb != nil {
				cb(newConfig.Settings())
			}
		}
	}
}

// applyLegacyConfigAliases 把旧版小驼峰 key 抄到 snake_case，现网 etcd 未改名时仍能加载。
func applyLegacyConfigAliases(v *viper.Viper) {
	if v == nil {
		return
	}
	if raw := v.Get("oss.buckets"); raw != nil {
		if m, err := oss.NormalizeBuckets(raw); err == nil && len(m) > 0 {
			v.Set("oss.buckets", m)
		}
	}
	for _, a := range []struct{ from, to string }{
		{"crosconf", "cros_conf"},
		{"upload.maxsizemb", "upload.max_size_mb"},
		{"upload.multipartmemorymb", "upload.multipart_memory_mb"},
		{"upload.allowedextensions", "upload.allowed_extensions"},
		{"upload.presignmaxsizemb", "upload.presign_max_size_mb"},
		{"upload.presignttl", "upload.presign_ttl"},
		{"upload.defaultscene", "upload.default_scene"},
		{"oss.fileurl", "oss.file_url"},
		{"oss.accesskey", "oss.access_key"},
		{"oss.secretkey", "oss.secret_key"},
		{"oss.pathstyle", "oss.path_style"},
	} {
		if v.IsSet(a.to) {
			continue
		}
		if v.IsSet(a.from) {
			v.Set(a.to, v.Get(a.from))
		}
	}
	applySceneCatalogAliases(v)
}

// applySceneCatalogAliases 兼容 upload.scenes.public: public 这种短写。
func applySceneCatalogAliases(v *viper.Viper) {
	raw := v.Get("upload.scenes")
	m, ok := raw.(map[string]any)
	if !ok || len(m) == 0 {
		return
	}
	out := make(map[string]any, len(m))
	changed := false
	for k, val := range m {
		if s, ok := val.(string); ok {
			out[k] = map[string]any{"visibility": s}
			changed = true
			continue
		}
		out[k] = val
	}
	if changed {
		v.Set("upload.scenes", out)
	}
}

// FileExists 相对路径按 resolveConfigFile 规则查找。
func FileExists(path string) bool {
	_, err := resolveConfigFile(path)
	return err == nil
}

// LoadInitProviderTemplate 读 init_default_settings.yaml：只含 oss 连接和桶，写入 oss_provider。
func LoadInitProviderTemplate(path string) (ProviderSpec, error) {
	resolved, err := resolveConfigFile(path)
	if err != nil {
		return ProviderSpec{}, err
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return ProviderSpec{}, fmt.Errorf("读取配置文件 %s: %w", resolved, err)
	}
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBuffer(raw)); err != nil {
		return ProviderSpec{}, fmt.Errorf("解析配置文件 %s: %w", resolved, err)
	}
	applyLegacyConfigAliases(v)
	var spec ProviderSpec
	if err := v.Unmarshal(&spec); err != nil {
		return ProviderSpec{}, fmt.Errorf("解析配置文件 %s: %w", resolved, err)
	}
	if err := spec.OSS.Validate(); err != nil {
		return ProviderSpec{}, fmt.Errorf("init template %s: %w", resolved, err)
	}
	return spec, nil
}

// Load 加载配置。path 非空读本地 YAML（不 watch）；空则读 etcd 并热更新。
func Load(path string) *AppConfig {
	once.Do(func() {
		inst = &AppConfig{}
		if strings.TrimSpace(path) != "" {
			resolved, err := inst.loadFromFile(path)
			if err != nil {
				panic(err)
			}
			log.Printf("配置来源: 本地文件 %s", resolved)
			return
		}
		if err := inst.initEtcdClient(); err != nil {
			panic(fmt.Sprintf("初始化 etcd 客户端失败: %v", err))
		}
		prefix := os.Getenv("ETCD_CONFIG_KEY")
		if prefix == "" {
			panic("ETCD_CONFIG_KEY 环境变量未设置")
		}
		etcdConfigKey = prefix + configPathSuffix
		inst.readConfig()
		log.Printf("配置来源: etcd %s", etcdConfigKey)
		go inst.watchConfig()
	})
	mu.RLock()
	defer mu.RUnlock()
	return inst
}

// GetInstance 默认走 etcd。本地文件请用 Load(path) 或 main 的 -config。
func GetInstance() *AppConfig {
	return Load("")
}

// Close 关闭 etcd 客户端。本地文件模式无需调用。
func Close() {
	if etcdClient != nil {
		if err := etcdClient.Close(); err != nil {
			log.Println(err)
		}
	}
}
