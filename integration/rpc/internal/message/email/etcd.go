package email

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

// watcher 监听一个邮件配置 Key，解析失败时保留最近一次有效配置。
type watcher struct {
	client      *clientv3.Client                     // 配置读取与监听共用的连接
	key         string                               // 完整配置 Key，不是前缀
	apply       func(config.EmailProvider, []string) // 兼容旧版单 SMTP 发布回调
	applyConfig func(config.EmailConf)               // 完整配置发布回调，优先使用
	ctx         context.Context                      // 控制读取、监听和重试的生命周期
	cancel      context.CancelFunc                   // 关闭时中止后台工作
}

const etcdWatchRetryDelay = time.Second // 读取失败或监听中断后的重试间隔

// etcdValue 兼容顶层和 email 嵌套两种配置格式，由 parseEtcdConfig 合并。
type etcdValue struct {
	Enabled         *bool                        `yaml:"enabled"`           // nil 表示未显式设置开关
	SMTPProvider    config.EmailProvider         `yaml:"smtp_provider"`     // 旧版单 SMTP 配置
	EmailBlacklist  []string                     `yaml:"email_blacklist"`   // 顶层域名黑名单
	Providers       []config.EmailProviderConfig `yaml:"providers"`         // 顶层多通道列表
	BlockChinaEmail bool                         `yaml:"block_china_email"` // 顶层内置域名拦截开关
	Retry           config.EmailRetryConfig      `yaml:"retry"`             // 通道重试和退避策略
	RecordLimit     int                          `yaml:"record_limit"`      // 内存记录和在途请求限额
	Email           struct {
		Enabled         *bool                        `yaml:"enabled"`           // 顶层未设置时使用
		SMTP            config.EmailProvider         `yaml:"smtp"`              // 顶层未填 SMTP 时使用
		Providers       []config.EmailProviderConfig `yaml:"providers"`         // 顶层列表为空时使用
		Blacklist       []string                     `yaml:"blacklist"`         // 顶层黑名单为空时使用
		BlockChinaEmail bool                         `yaml:"block_china_email"` // 与顶层开关取启用值
	} `yaml:"email"` // 嵌套邮件配置
}

// newWatcher 为只接收单 SMTP 配置的调用方保留兼容入口。
func newWatcher(c config.EmailEtcdConfig, apply func(config.EmailProvider, []string)) (*watcher, error) {
	return newWatcherWithConfig(c, apply, nil)
}

// newWatcherWithConfig 创建单 Key 监听器；由 run 在后台读取并发布配置。
func newWatcherWithConfig(c config.EmailEtcdConfig, apply func(config.EmailProvider, []string), applyConfig func(config.EmailConf)) (*watcher, error) {
	// 清理空节点；没有可用地址时拒绝创建，实际读取由后台执行。
	hosts := make([]string, 0, len(c.Hosts))
	for _, host := range c.Hosts {
		if host = strings.TrimSpace(host); host != "" {
			hosts = append(hosts, host)
		}
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("email etcd hosts are required")
	}
	// 未指定时沿用旧版 SMTP Key，内容仍可使用多通道配置格式。
	key := strings.TrimSpace(c.Key)
	if key == "" {
		key = defaultEtcdKey
	}
	timeout := defaultEtcdKeyTimeout(c.DialTimeout)
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,   // 去除空项后的 Etcd 节点列表
		DialTimeout: timeout, // 仅控制建立连接的超时
	})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{
		client:      client,      // 复用读取与监听连接
		key:         key,         // 指定值或默认配置 Key
		apply:       apply,       // 旧版单通道回调
		applyConfig: applyConfig, // 完整配置回调
		ctx:         ctx,         // 控制后台监听退出
		cancel:      cancel,      // 释放时取消读取与监听
	}, nil
}

// run 先读取配置，再从读取版本的下一版本监听，避免 Get 和 Watch 之间遗漏更新。
// 监听中断后重新读取，以处理断线期间的配置变更或删除。
func (w *watcher) run() {
	// 首次读取失败时每秒重试，服务关闭会中止等待。
	revision, loaded := w.load()
	for !loaded {
		if !w.waitForRetry() {
			return
		}
		revision, loaded = w.load()
	}
	for {
		// 从快照下一修订版本接续监听，兼顾读取期间发生的更新。
		watch := w.client.Watch(w.ctx, w.key, clientv3.WithRev(revision+1))
		watchClosed := false
		for response := range watch {
			if err := response.Err(); err != nil {
				// 历史版本被压缩或监听异常时退出本轮，随后重新读取当前配置。
				if errors.Is(err, rpctypes.ErrCompacted) {
					logx.Errorf("email etcd watch compacted key=%s: %v", w.key, err)
				} else if w.ctx.Err() == nil {
					logx.Errorf("email etcd watch %s: %v", w.key, err)
				}
				watchClosed = true
				break
			}
			revision = response.Header.Revision
			for _, event := range response.Events {
				w.handleEvent(event)
			}
		}
		if w.ctx.Err() != nil {
			return
		}
		if !watchClosed {
			logx.Errorf("email etcd watch closed unexpectedly key=%s", w.key)
		}
		// 断线期间 Key 可能已删除，重读时须同时处理缺失和配置变更。
		if latest, ok := w.loadAndClearMissing(); ok {
			revision = latest
		}
		if !w.waitForRetry() {
			return
		}
	}
}

// handleEvent 发布有效配置；删除 Key 时停用通道，无效内容保留原配置。
func (w *watcher) handleEvent(event *clientv3.Event) {
	if event == nil || event.Kv == nil {
		return
	}
	switch event.Type {
	case clientv3.EventTypePut:
		// 解析有效配置后发布；无效文档不覆盖正在使用的快照。
		cfg, err := parseEtcdConfig(event.Kv.Value)
		if err != nil {
			logx.Errorf("email etcd parse %s: %v", w.key, err)
			return
		}
		if w.applyConfig != nil {
			w.applyConfig(cfg)
		} else if w.apply != nil {
			w.apply(cfg.Provider, cfg.Blacklist)
		}
	case clientv3.EventTypeDelete:
		logx.Infof("email etcd config deleted key=%s; disabling provider", w.key)
		if w.applyConfig != nil {
			w.applyConfig(config.EmailConf{})
		} else if w.apply != nil {
			w.apply(config.EmailProvider{}, nil)
		}
	}
}

// load 执行首次读取，返回已读取版本和是否完成读取；Key 缺失也视为读取成功。
func (w *watcher) load() (int64, bool) {
	return w.loadAndClearMissing()
}

// loadAndClearMissing 将 Key 缺失视为禁用；读取失败或内容无效则保留现有配置。
// 启动读取和断线重载使用相同的缺失语义。
func (w *watcher) loadAndClearMissing() (int64, bool) {
	ctx, cancel := context.WithTimeout(w.ctx, defaultSendTimeout)
	defer cancel()
	response, err := w.client.Get(ctx, w.key)
	if err != nil {
		if w.ctx.Err() == nil {
			logx.Errorf("email etcd reload %s: %v", w.key, err)
		}
		return 0, false
	}
	if len(response.Kvs) == 0 {
		// Etcd 是配置来源时，Key 缺失表示停用，不回退到本地账号继续发送。
		logx.Infof("email etcd config missing key=%s; disabling provider", w.key)
		if w.applyConfig != nil {
			w.applyConfig(config.EmailConf{})
		} else if w.apply != nil {
			w.apply(config.EmailProvider{}, nil)
		}
		return response.Header.Revision, true
	}
	cfg, err := parseEtcdConfig(response.Kvs[0].Value)
	if err != nil {
		logx.Errorf("email etcd parse %s after watch reset: %v", w.key, err)
		return response.Header.Revision, true
	}
	if w.applyConfig != nil {
		w.applyConfig(cfg)
	} else if w.apply != nil {
		w.apply(cfg.Provider, cfg.Blacklist)
	}
	return response.Header.Revision, true
}

// waitForRetry 每次失败后等待 1 秒，可持续重试，关闭服务时立即结束等待。
func (w *watcher) waitForRetry() bool {
	timer := time.NewTimer(etcdWatchRetryDelay)
	defer timer.Stop()
	select {
	case <-w.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// close 取消后台工作并关闭 Etcd 连接，协程退出由 Service.Close 等待。
func (w *watcher) close() {
	w.cancel()
	_ = w.client.Close()
}

// parseEtcd 为旧版调用方提取单 SMTP 配置；未填单通道时取首个已启用的 SMTP。
func parseEtcd(data []byte) (config.EmailProvider, []string, error) {
	cfg, err := parseEtcdConfig(data)
	if err != nil {
		return config.EmailProvider{}, nil, err
	}
	provider := cfg.Provider
	if !hasProvider(provider) {
		for _, p := range cfg.Providers {
			if strings.EqualFold(p.Type, "smtp") && p.Enabled {
				provider = p.SMTP
				break
			}
		}
	}
	return provider, cfg.Blacklist, nil
}

// parseEtcdConfig 接收单 SMTP、多通道及 email 嵌套格式，只接收单个 YAML 文档。
// 空内容、未知字段和无已知配置字段的文档均拒绝，避免拼写错误清空有效设置。
func parseEtcdConfig(data []byte) (config.EmailConf, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return config.EmailConf{}, fmt.Errorf("email configuration cannot be empty")
	}
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&document); err != nil {
		return config.EmailConf{}, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return config.EmailConf{}, fmt.Errorf("email configuration must contain one YAML document")
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return config.EmailConf{}, fmt.Errorf("email configuration must be a mapping")
	}
	root := document.Content[0]
	// 至少包含一个已知根字段，防止无关 YAML 被误当成空配置。
	recognized := false
	for index := 0; index < len(root.Content); index += 2 {
		switch root.Content[index].Value {
		case "enabled", "smtp_provider", "email_blacklist", "providers", "block_china_email", "retry", "record_limit", "email":
			recognized = true
		}
	}
	if !recognized {
		return config.EmailConf{}, fmt.Errorf("email configuration contains no recognized fields")
	}
	var value etcdValue
	// Node.Decode 不执行 KnownFields 检查，因此使用严格解码器解析具体配置类型。
	decoder = yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&value); err != nil {
		return config.EmailConf{}, err
	}

	cfg := config.EmailConf{
		Providers:       value.Providers,       // 顶层多通道配置
		Blacklist:       value.EmailBlacklist,  // 顶层域名黑名单
		BlockChinaEmail: value.BlockChinaEmail, // 内置邮箱域名拦截开关
		Retry:           value.Retry,           // 通道重试策略
		RecordLimit:     value.RecordLimit,     // 内存记录与并发请求限额
	}
	// 顶层列表为空时兼容 email 嵌套格式，内置域名开关合并两层的启用值。
	if len(cfg.Providers) == 0 {
		cfg.Providers = value.Email.Providers
	}
	if len(cfg.Blacklist) == 0 {
		cfg.Blacklist = value.Email.Blacklist
	}
	if !cfg.BlockChinaEmail {
		cfg.BlockChinaEmail = value.Email.BlockChinaEmail
	}
	// 顶层 SMTP 和开关优先；仅在顶层未配置时兼容 email 内的旧字段。
	provider := normalizeSMTPProvider(value.SMTPProvider)
	enabled := value.Enabled
	if enabled == nil {
		enabled = value.Email.Enabled
	}
	if !smtpConfigured(provider) {
		provider = normalizeSMTPProvider(value.Email.SMTP)
	}
	if hasProvider(provider) || provider.Host != "" || provider.AuthEmail != "" || provider.AuthPassword != "" {
		if strings.TrimSpace(provider.TLSMode) == "" {
			provider.TLSMode = "implicit"
		}
		if enabled == nil || *enabled {
			if err := validateProvider(provider); err != nil {
				return config.EmailConf{}, err
			}
		}
		cfg.Provider = provider
	}
	// 未显式设置开关时按可用配置推断；显式值覆盖推断结果，最后统一校验。
	cfg.Enabled = hasProvider(cfg.Provider) || enabledProviders(cfg.Providers)
	if value.Enabled != nil {
		cfg.Enabled = *value.Enabled
	} else if value.Email.Enabled != nil {
		cfg.Enabled = *value.Email.Enabled
	}
	normalized, err := buildSnapshot(cfg)
	if err != nil {
		return config.EmailConf{}, err
	}
	cfg.Providers = normalized.Providers
	return cfg, nil
}

// smtpConfigured 判断是否填写了任一 SMTP 账号字段，用于识别未填完整的配置。
func smtpConfigured(provider config.EmailProvider) bool {
	return strings.TrimSpace(provider.Host) != "" || strings.TrimSpace(provider.AuthEmail) != "" || strings.TrimSpace(provider.AuthPassword) != ""
}

// defaultEtcdKeyTimeout 换算 Etcd 连接超时，未配置时使用 5 秒；不控制 Get 超时。
func defaultEtcdKeyTimeout(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultDialTimeout
	}
	return time.Duration(seconds) * time.Second
}
