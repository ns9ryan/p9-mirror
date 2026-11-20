// Package oss 对象存储驱动（S3 / MinIO）。只做对象存取，不管配置从哪来。
// HTTP 上传直接调本包；档案入库走 rpc。调用方注入 Config，热更新时再 Apply。
// 建桶、访问策略、CDN 由运维配置。对接三方时实现 Store；直传再实现 PresignStore。
package oss

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrNotFound 对象不存在（Head / Exists）。
var ErrNotFound = errors.New("oss: object not found")

// ErrPresignUnsupported 当前驱动未实现 PresignStore（浏览器直传）。
var ErrPresignUnsupported = errors.New("oss: driver does not support presign")

// Store 对象存储抽象。只负责存取；建桶与访问策略由运维配置。
// 新驱动至少实现本接口。浏览器直传再额外实现 PresignStore。
type Store interface {
	PutObject(ctx context.Context, in PutObjectInput) error
	ObjectExists(ctx context.Context, bucket, key string) (bool, error)
	// ObjectURL 返回访问地址。public 桶按 path_style 拼 file_url；private 桶预签名 GET。
	// ttl 仅对 private 有效：<=0 用桶配置的 sign_ttl；>0 覆盖有效期。public 忽略 ttl。
	ObjectURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error)
}

// PresignStore 预签名直传。S3/MinIO 已实现；其它驱动可不实现。
// 对应 HTTP：签发 PUT、complete 时 Head、超限 Delete。
type PresignStore interface {
	PresignPut(ctx context.Context, in PresignPutInput) (PresignPutResult, error)
	HeadObject(ctx context.Context, bucket, key string) (ObjectHead, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

// PresignPutInput 预签名 PUT。TTL 由调用方决定。
// ContentType、ContentLength（>0 时）会签进 URL，客户端 PUT 必须原样带上。
type PresignPutInput struct {
	Bucket        string
	Key           string
	ContentType   string
	ContentLength int64
	TTL           time.Duration
}

type PresignPutResult struct {
	URL     string
	Method  string
	Headers map[string]string
}

type ObjectHead struct {
	Size        int64
	ContentType string
}

type PutObjectInput struct {
	Bucket        string
	Key           string
	Body          io.ReadSeeker
	ContentLength int64
	ContentType   string
}

// PutInput 业务侧上传入参；Body 可为 ReadSeeker 或 *bytes.Buffer。
type PutInput struct {
	Bucket        string
	Key           string
	Body          io.Reader
	ContentLength int64
	ContentType   string
}
