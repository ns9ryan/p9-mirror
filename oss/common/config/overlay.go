package config

import (
	"oa.98ent.com/p9/oss/common/oss"
)

// ProviderSpec 后台服务商行：只覆盖 oss 连接和桶。scene 目录仍用进程 yaml/etcd。
type ProviderSpec struct {
	OSS oss.Config `mapstructure:"oss"`
}

// CloneSettings 深拷贝运行时配置（含 map）。
func CloneSettings(s Settings) Settings {
	out := s
	out.Upload.AllowedExtensions = append([]string(nil), s.Upload.AllowedExtensions...)
	out.Upload.Scenes = cloneSceneMap(s.Upload.Scenes)
	out.OSS.Buckets = cloneBucketMap(s.OSS.Buckets)
	return out
}

// ApplyProvider 用服务商覆盖 base 的 oss，再按 visibility 重新对齐桶。上传限制与 scene 目录不变。
func ApplyProvider(base Settings, spec ProviderSpec) (Settings, error) {
	s := CloneSettings(base)
	s.OSS = spec.OSS
	s.OSS.Buckets = cloneBucketMap(spec.OSS.Buckets)
	if s.OSS.Driver == "" {
		s.OSS.Driver = "s3"
	}
	if s.OSS.Region == "" {
		s.OSS.Region = "default"
	}
	s.bindScenes()
	if err := s.Validate(); err != nil {
		return Settings{}, err
	}
	return s, nil
}

func cloneSceneMap(src map[string]SceneSettings) map[string]SceneSettings {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]SceneSettings, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneBucketMap(src map[string]oss.Bucket) map[string]oss.Bucket {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]oss.Bucket, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// CloneProviderSpec 深拷贝服务商模板。
func CloneProviderSpec(s ProviderSpec) ProviderSpec {
	out := s
	out.OSS.Buckets = cloneBucketMap(s.OSS.Buckets)
	return out
}
