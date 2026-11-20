package oss

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Client 对象存储客户端。由调用方注入 Config 创建，本包不加载、不 watch 配置。
// 热更新时调用方自行 Apply 新配置；Put / Exists / URL 与 Apply 可并发。
type Client struct {
	mu   sync.RWMutex
	cfg  Config
	impl Store
}

// New 校验配置并建立存储连接。cfg 从何处来（etcd / 环境变量 / 文件）由调用方决定。
func New(cfg Config) (*Client, error) {
	c := &Client{}
	if err := c.Apply(cfg); err != nil {
		return nil, err
	}
	return c, nil
}

// Apply 用新配置重建内部 Store。校验失败时保持原连接不变。
func (c *Client) Apply(cfg Config) error {
	if c == nil {
		return fmt.Errorf("oss: client is nil")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	impl, err := NewStore(cfg)
	if err != nil {
		return err
	}
	cfg.Buckets = cloneBuckets(cfg.Buckets)
	c.mu.Lock()
	c.cfg = cfg
	c.impl = impl
	c.mu.Unlock()
	return nil
}

func (c *Client) store() (Store, error) {
	if c == nil {
		return nil, fmt.Errorf("oss: client is nil")
	}
	c.mu.RLock()
	impl := c.impl
	c.mu.RUnlock()
	if impl == nil {
		return nil, fmt.Errorf("oss: store is not ready")
	}
	return impl, nil
}

func (c *Client) presignStore() (PresignStore, error) {
	s, err := c.store()
	if err != nil {
		return nil, err
	}
	p, ok := s.(PresignStore)
	if !ok {
		return nil, fmt.Errorf("%w", ErrPresignUnsupported)
	}
	return p, nil
}

// Config 返回当前配置的拷贝。
func (c *Client) Config() Config {
	if c == nil {
		return Config{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := c.cfg
	out.Buckets = cloneBuckets(c.cfg.Buckets)
	return out
}

// Put 上传对象并返回可访问 URL（按桶 mode：public 公开链，private 预签名 GET）。Body 会适配为 ReadSeeker。
func (c *Client) Put(ctx context.Context, in PutInput) (string, error) {
	s, err := c.store()
	if err != nil {
		return "", err
	}
	body, size, err := toReadSeeker(in.Body, in.ContentLength)
	if err != nil {
		return "", err
	}
	if err := s.PutObject(ctx, PutObjectInput{
		Bucket:        in.Bucket,
		Key:           in.Key,
		Body:          body,
		ContentLength: size,
		ContentType:   in.ContentType,
	}); err != nil {
		return "", err
	}
	return s.ObjectURL(ctx, in.Bucket, in.Key, 0)
}

// Exists 判断对象是否存在。
func (c *Client) Exists(ctx context.Context, bucket, key string) (bool, error) {
	s, err := c.store()
	if err != nil {
		return false, err
	}
	return s.ObjectExists(ctx, bucket, key)
}

// URL 返回对象访问地址。形态由桶 mode 决定，调用方不必区分公有/私有：
// public 按 path_style 拼 file_url（true：/{bucket}/{key}，false：/{key}）；
// private → 预签名 GET（默认桶 sign_ttl）。可选 ttl 只覆盖 private；public 忽略 ttl。
func (c *Client) URL(ctx context.Context, bucket, key string, ttl ...time.Duration) (string, error) {
	s, err := c.store()
	if err != nil {
		return "", err
	}
	var d time.Duration
	if len(ttl) > 0 {
		d = ttl[0]
	}
	return s.ObjectURL(ctx, bucket, key, d)
}

// PresignPut 签发浏览器直传用的预签名 PUT。本包不经手文件字节。
// ContentLength > 0 时签进 URL，客户端 PUT 的 Content-Length 必须与之一致。
func (c *Client) PresignPut(ctx context.Context, in PresignPutInput) (PresignPutResult, error) {
	s, err := c.presignStore()
	if err != nil {
		return PresignPutResult{}, err
	}
	return s.PresignPut(ctx, in)
}

// Head 返回对象大小与 Content-Type；不存在时返回错误。驱动须实现 PresignStore。
func (c *Client) Head(ctx context.Context, bucket, key string) (ObjectHead, error) {
	s, err := c.presignStore()
	if err != nil {
		return ObjectHead{}, err
	}
	return s.HeadObject(ctx, bucket, key)
}

// Delete 删除对象。驱动须实现 PresignStore。
func (c *Client) Delete(ctx context.Context, bucket, key string) error {
	s, err := c.presignStore()
	if err != nil {
		return err
	}
	return s.DeleteObject(ctx, bucket, key)
}
