package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"oa.98ent.com/p9/oss/common/oss"
)

// DiffSettings 对比热更新前后的运行时配置，返回运维可读的变更行（密钥不打明文）。
func DiffSettings(prev, next Settings) []string {
	var lines []string
	add := func(name, from, to string) {
		if from == to {
			return
		}
		lines = append(lines, fmt.Sprintf("%s: %s -> %s", name, from, to))
	}

	add("port", quote(prev.Port), quote(next.Port))
	if prev.Port != next.Port {
		lines[len(lines)-1] += " （监听端口需重启进程才生效）"
	}

	add("upload.max_size_mb", mb(prev.Upload.MaxSize), mb(next.Upload.MaxSize))
	add("upload.multipart_memory_mb", mb(prev.Upload.MultipartMemory), mb(next.Upload.MultipartMemory))
	add("upload.presign_max_size_mb", mb(prev.Upload.PresignMaxSize), mb(next.Upload.PresignMaxSize))
	add("upload.presign_ttl", seconds(prev.Upload.PresignTTL), seconds(next.Upload.PresignTTL))
	add("upload.allowed_extensions", joinExt(prev.Upload.AllowedExtensions), joinExt(next.Upload.AllowedExtensions))
	add("upload.default_scene", quote(prev.Upload.DefaultScene), quote(next.Upload.DefaultScene))
	add("upload.scenes", joinScenes(prev.Upload.Scenes), joinScenes(next.Upload.Scenes))

	add("oss.driver", prev.OSS.Driver, next.OSS.Driver)
	add("oss.endpoint", quote(prev.OSS.Endpoint), quote(next.OSS.Endpoint))
	add("oss.file_url", quote(prev.OSS.FileURL), quote(next.OSS.FileURL))
	add("oss.region", prev.OSS.Region, next.OSS.Region)
	add("oss.path_style", strconv.FormatBool(prev.OSS.PathStyle), strconv.FormatBool(next.OSS.PathStyle))
	add("oss.buckets", joinBuckets(prev.OSS.Buckets), joinBuckets(next.OSS.Buckets))
	addSecret("oss.access_key", prev.OSS.AccessKey, next.OSS.AccessKey, &lines)
	addSecret("oss.secret_key", prev.OSS.SecretKey, next.OSS.SecretKey, &lines)
	return lines
}

// addSecret 密钥变更只记「已变更」，不打印明文。
func addSecret(name, from, to string, lines *[]string) {
	if from == to {
		return
	}
	*lines = append(*lines, name+": 已变更（不打印明文）")
}

// quote 空字符串显示为 (空)，便于对比日志。
func quote(s string) string {
	if s == "" {
		return "(空)"
	}
	return s
}

// joinList 把字符串列表拼成对比日志。
func joinList(ss []string) string {
	if len(ss) == 0 {
		return "(空)"
	}
	return strings.Join(ss, ",")
}

// joinBuckets 把桶 map 格式化成 name:mode@ttl，便于热更新日志。
func joinBuckets(m map[string]oss.Bucket) string {
	if len(m) == 0 {
		return "(空)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		b := m[k]
		part := k + ":" + b.Mode
		if b.SignTTL > 0 {
			part += fmt.Sprintf("@%ds", b.SignTTL)
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ",")
}

// joinExt 格式化后缀白名单；空表示全部允许。
func joinExt(ss []string) string {
	if len(ss) == 0 {
		return "(空=全部允许)"
	}
	return strings.Join(ss, ",")
}

// joinScenes 格式化 scene:visibility@bucket 列表。
func joinScenes(m map[string]SceneSettings) string {
	if len(m) == 0 {
		return "(空)"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		sc := m[k]
		parts = append(parts, fmt.Sprintf("%s:%s@%s", k, sc.Visibility, sc.Bucket))
	}
	return strings.Join(parts, ",")
}

// mb 把字节转成 MB 字符串。
func mb(n int64) string {
	return strconv.FormatInt(n>>20, 10)
}

// seconds 把 Duration 转成秒数字符串。
func seconds(d time.Duration) string {
	return strconv.FormatInt(int64(d.Seconds()), 10) + "s"
}
