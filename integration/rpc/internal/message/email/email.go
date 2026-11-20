// Package email 提供邮件校验、通道选择、发送重试和内存投递记录。
package email

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrNotConfigured       = errors.New("email provider not configured")                              // 服务关闭或没有可用通道
	ErrInvalidRequest      = errors.New("invalid email request")                                      // 正文、主题或收件人数不符合限制
	ErrInvalidRecipient    = errors.New("invalid email recipient")                                    // 收件地址格式无效
	ErrRecipientBlocked    = errors.New("email recipient blocked")                                    // 收件域名被拦截
	ErrInvalidAttachment   = errors.New("invalid email attachment")                                   // 附件格式或大小不符合限制
	ErrProviderUnavailable = errors.New("email provider unavailable")                                 // 可按策略重试的不可用错误
	ErrProviderUnsupported = errors.New("email provider does not support this request")               // 当前通道不支持，可尝试下一通道
	ErrProviderRejected    = errors.New("email provider rejected the request")                        // 明确拒绝，不在当前通道重试
	ErrDeliveryUncertain   = errors.New("email delivery status unknown; do not automatically resend") // 受理不明，停止重发
)

const (
	defaultEtcdKey      = "/integration/message/email/providers/smtp" // 未指定时读取的配置 Key
	defaultDialTimeout  = 5 * time.Second                             // 默认连接超时
	defaultSendTimeout  = 10 * time.Second                            // 默认发送超时，也用于邮件配置读取
	defaultImplicitPort = "465"                                       // 连接即启用 TLS 的默认端口
	defaultStartTLSPort = "587"                                       // STARTTLS 升级的默认端口
)

// Attachment 使用 Base64 接收附件内容，构建 MIME 邮件时解码并检查大小。
type Attachment struct {
	Filename string // 附件文件名
	Base64   string // 标准 Base64 内容，不带 data URL 前缀
	MIME     string // 媒体类型，如 application/pdf
}

// Request 是邮件业务请求，不依赖 RPC 生成类型，供业务层统一调用。
type Request struct {
	Recipients  []string     // 主收件人
	CC          []string     // 抄送，写入邮件头
	BCC         []string     // 密送，仅用于投递
	Subject     string       // 邮件主题
	Content     string       // 纯文本正文，与 HTML 至少填写一项
	HTMLContent string       // HTML 正文，可与纯文本并存
	Attachments []Attachment // 附件，解码后总量不超过 10 MiB
	Provider    string       // 空值按优先级选择，否则先匹配名称再匹配类型
}

// SendResult 表示通道受理结果，不代表邮件已进入收件箱。
type SendResult struct {
	Accepted          int    // 去重后的受理人数，包含抄送和密送
	Provider          string // 实际使用的通道名称
	DeliveryID        string // 本服务内存记录编号，不是幂等键
	ProviderMessageID string // 厂商回执编号，SMTP 通道为空
}

// DeliveryRecord 仅在当前进程内保存发送状态和排查信息，不保存邮件正文。
type DeliveryRecord struct {
	ID                string    // 本服务生成的投递编号
	Provider          string    // 最后尝试的通道
	Status            string    // pending 处理中、sent 已受理、failed 失败、unknown 结果不确定
	Accepted          int       // 厂商受理人数
	Attempts          int       // 所有通道累计尝试次数
	Error             string    // 最后一次错误
	ProviderMessageID string    // 厂商回执编号，可能为空
	CreatedAt         time.Time // 开始记录时间
	UpdatedAt         time.Time // 最终状态更新时间
}

// snapshot 发布后只读；一次发送固定使用同一份配置，避免热更新混用通道参数。
type snapshot struct {
	Provider        config.EmailProvider         // 兼容旧版单 SMTP 配置
	Providers       []config.EmailProviderConfig // 多通道配置，优先于单 SMTP
	Blacklist       []string                     // 规范化后的域名黑名单
	BlockChinaEmail bool                         // 是否启用内置邮箱域名拦截
	Retry           config.EmailRetryConfig      // 已补默认值的重试策略
	RecordLimit     int                          // 完成记录和在途请求分别使用的限额
	Enabled         bool                         // 本地总开关与远程配置共同决定的有效状态
}

// Service 编排邮件发送，通过原子快照读取配置，通过互斥锁维护投递记录。
type Service struct {
	enabled   bool                     // 本地总开关，远程配置不能覆盖
	closed    atomic.Bool              // 服务是否已关闭
	current   atomic.Pointer[snapshot] // 当前只读配置快照
	etcd      *watcher                 // 可选的远程配置监听器
	wg        sync.WaitGroup           // 等待配置监听退出
	ctx       context.Context          // 服务生命周期，用于取消在途发送
	cancel    context.CancelFunc       // 关闭服务时触发取消
	closeOnce sync.Once                // 保证资源只释放一次
	recordMu  sync.Mutex               // 保护内存投递记录
	records   []DeliveryRecord         // 在途和完成记录，不写数据库

	send func(context.Context, config.EmailProviderConfig, []string, bool, Request) (SendResult, error) // 可注入的发送函数
}

// New 校验本地配置并启动可选的 Etcd 监听；远程配置在后台加载。
func New(c config.EmailConf) (*Service, error) {
	cfg, err := buildSnapshot(c)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		enabled: c.Enabled, // 保留本地总开关
		ctx:     ctx,       // 统一控制发送取消
		cancel:  cancel,    // 服务关闭时释放
	}
	s.current.Store(cfg)
	if c.Enabled && len(c.Etcd.Hosts) != 0 {
		w, err := newWatcherWithConfig(c.Etcd, s.apply, s.applyConfig)
		if err != nil {
			cancel()
			return nil, err
		}
		s.etcd = w
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			w.run()
		}()
	}
	return s, nil
}

// normalizeRetry 补齐默认退避参数，并限制尝试次数和等待上限。
func normalizeRetry(r config.EmailRetryConfig) config.EmailRetryConfig {
	// 每通道默认最多尝试 3 次（含首次），最多允许 10 次。
	if r.MaxAttempts < 1 {
		r.MaxAttempts = 3
	}
	if r.MaxAttempts > 10 {
		r.MaxAttempts = 10
	}
	// 首次退避默认 100 毫秒，上限默认 5 秒，配置最大不超过 60 秒。
	if r.BackoffMS < 0 {
		r.BackoffMS = 0
	}
	if r.BackoffMS == 0 {
		r.BackoffMS = 100
	}
	if r.MaxBackoffMS <= 0 {
		r.MaxBackoffMS = 5000
	}
	if r.MaxBackoffMS > 60000 {
		r.MaxBackoffMS = 60000
	}
	if r.BackoffMS > r.MaxBackoffMS {
		r.BackoffMS = r.MaxBackoffMS
	}
	return r
}

// Close 停止配置监听并取消正在执行的发送，可重复调用。
func (s *Service) Close() error {
	if s == nil {
		return nil
	}
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		if s.cancel != nil {
			s.cancel()
		}
		if s.etcd != nil {
			s.etcd.close()
			s.wg.Wait()
		}
	})
	return nil
}

// Send 为仅关心受理人数的调用方保留简化入口。
func (s *Service) Send(ctx context.Context, req Request) (int, error) {
	result, err := s.SendResult(ctx, req)
	return result.Accepted, err
}

// SendResult 校验地址和正文后按通道发送；附件在 MIME 构建时校验，结果仅记录在内存。
func (s *Service) SendResult(ctx context.Context, req Request) (SendResult, error) {
	if s == nil || s.closed.Load() || s.ctx == nil {
		return SendResult{}, ErrNotConfigured
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}
	// 请求取消或服务关闭都终止发送；本批固定读取一份配置快照。
	requestCtx, cancel := context.WithCancel(ctx)
	stopCancel := context.AfterFunc(s.ctx, cancel)
	defer stopCancel()
	defer cancel()
	ctx = requestCtx
	cfg := s.current.Load()
	if cfg == nil || !cfg.Enabled {
		return SendResult{}, ErrNotConfigured
	}
	id := newDeliveryID()
	rec := DeliveryRecord{
		ID:        id,         // 本次投递编号，重试沿用
		Status:    "pending",  // 发送前先登记处理中
		CreatedAt: time.Now(), // 请求开始时间
		UpdatedAt: time.Now(), // 初始状态时间
	}
	// 发送前占用记录容量；容量不足时直接拒绝，避免发出后无法追踪状态。
	if err := s.storeRecord(rec); err != nil {
		return SendResult{}, err
	}
	// 校验并去重收件地址，再按名称或优先级选择可用通道。
	clean, _, err := validateRequest(req, cfg.Blacklist, cfg.BlockChinaEmail)
	if err != nil {
		s.finishRecord(id, "", "failed", "", 0, 0, err)
		return SendResult{DeliveryID: id}, err
	}
	req = clean
	providers := s.providerList(cfg, req.Provider)
	if len(providers) == 0 {
		s.finishRecord(id, "", "failed", "", 0, 0, ErrNotConfigured)
		return SendResult{DeliveryID: id}, ErrNotConfigured
	}
	var last error
	lastProvider := ""
	attempts := 0
	sender := s.send
	if sender == nil {
		sender = sendProvider
	}
	for _, p := range providers {
		lastProvider = p.Name
		max := cfg.Retry.MaxAttempts
		if max < 1 {
			max = 1
		}
		for attempt := 1; attempt <= max; attempt++ {
			if err := ctx.Err(); err != nil {
				s.finishRecord(id, p.Name, "failed", "", 0, attempts, err)
				return SendResult{DeliveryID: id}, err
			}
			attempts++
			result, err := sender(ctx, p, cfg.Blacklist, cfg.BlockChinaEmail, req)
			if err == nil {
				if result.ProviderMessageID == "" {
					result.ProviderMessageID = result.DeliveryID
				}
				result.DeliveryID = id
				result.Provider = p.Name
				s.finishRecord(id, p.Name, "sent", result.ProviderMessageID, result.Accepted, attempts, nil)
				return result, nil
			}
			last = err
			// 可能已经提交成功，立即停止所有重试和通道切换，防止重复发信。
			if errors.Is(err, ErrDeliveryUncertain) {
				s.finishRecord(id, p.Name, "unknown", "", 0, attempts, err)
				return SendResult{DeliveryID: id, Provider: p.Name}, err
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				s.finishRecord(id, p.Name, "failed", "", 0, attempts, err)
				return SendResult{DeliveryID: id}, err
			}
			// 当前通道明确无法处理时跳到下一通道，不在同一通道重复尝试。
			if errors.Is(err, ErrProviderUnsupported) || errors.Is(err, ErrProviderRejected) {
				break
			}
			// 内容、收件人或附件校验失败时停止整次发送，不通过切换通道绕过限制。
			if errors.Is(err, ErrInvalidRequest) || errors.Is(err, ErrInvalidRecipient) || errors.Is(err, ErrRecipientBlocked) || errors.Is(err, ErrInvalidAttachment) {
				s.finishRecord(id, p.Name, "failed", "", 0, attempts, err)
				return SendResult{DeliveryID: id}, err
			}
			if !errors.Is(err, ErrProviderUnavailable) {
				s.finishRecord(id, p.Name, "failed", "", 0, attempts, err)
				return SendResult{DeliveryID: id}, err
			}
			// 只有明确可重试的不可用错误才退避重试，等待也受请求取消控制。
			if attempt < max {
				backoff := time.Duration(cfg.Retry.BackoffMS*(1<<(attempt-1))) * time.Millisecond
				if backoff > time.Duration(cfg.Retry.MaxBackoffMS)*time.Millisecond {
					backoff = time.Duration(cfg.Retry.MaxBackoffMS) * time.Millisecond
				}
				select {
				case <-ctx.Done():
					s.finishRecord(id, p.Name, "failed", "", 0, attempts, ctx.Err())
					return SendResult{DeliveryID: id}, ctx.Err()
				case <-time.After(backoff):
				}
			}
		}
	}
	s.finishRecord(id, lastProvider, "failed", "", 0, attempts, last)
	return SendResult{DeliveryID: id}, last
}

// providerList 优先精确匹配通道名称，否则按类型筛选，并按 Priority 升序排列。
// 仅未配置多通道列表时使用单 SMTP 配置；指定已禁用的名称不会回退。
func (s *Service) providerList(cfg *snapshot, selected string) []config.EmailProviderConfig {
	selected = strings.TrimSpace(selected)
	legacy := len(cfg.Providers) == 0 && hasProvider(cfg.Provider)
	// 复制列表后再筛选排序，避免修改供其他发送请求共用的配置快照。
	out := append([]config.EmailProviderConfig(nil), cfg.Providers...)
	if legacy {
		out = []config.EmailProviderConfig{{
			Name:     "smtp",       // 旧版单通道的默认名称
			Type:     "smtp",       // 使用 SMTP 发送
			Enabled:  true,         // 沿用已通过总开关检查的旧配置
			Priority: 0,            // 旧版只有一个通道
			SMTP:     cfg.Provider, // 旧版 SMTP 账号配置
		}}
	}
	if selected != "" {
		// 名称精确命中后直接返回；命中禁用通道也不再按类型寻找替代项。
		for _, provider := range out {
			if provider.Name == selected {
				if provider.Enabled {
					return []config.EmailProviderConfig{provider}
				}
				return nil
			}
		}
		// 名称未命中才按类型筛选，ses 兼容为 aws 类型。
		filtered := out[:0]
		for _, p := range out {
			if strings.EqualFold(p.Type, selected) || (p.Type == "aws" && strings.EqualFold(selected, "ses")) {
				filtered = append(filtered, p)
			}
		}
		out = filtered
	}
	// 数值越小越先尝试；相同优先级保持原配置顺序。
	sort.SliceStable(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	if !legacy {
		enabled := out[:0]
		for _, p := range out {
			if p.Enabled {
				enabled = append(enabled, p)
			}
		}
		out = enabled
	}
	return out
}

// enabledProviders 判断列表中是否至少有一个启用的通道，不在此校验账号。
func enabledProviders(providers []config.EmailProviderConfig) bool {
	for _, p := range providers {
		if p.Enabled {
			return true
		}
	}
	return false
}

// normalizeRecordLimit 默认保留 100 条完成记录，最多 10000 条；在途请求另用相同限额。
func normalizeRecordLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 10000 {
		return 10000
	}
	return limit
}

// sendProvider 负责厂商分发；重试和切换统一由 Service 控制。
func sendProvider(ctx context.Context, p config.EmailProviderConfig, blacklist []string, blockChina bool, req Request) (SendResult, error) {
	switch strings.ToLower(p.Type) {
	case "", "smtp":
		accepted, err := sendSMTP(ctx, p.SMTP, blacklist, blockChina, req)
		return SendResult{Accepted: accepted, Provider: p.Name}, err
	case "aliyun":
		return sendAliyun(ctx, p, blacklist, blockChina, req)
	case "aws", "ses":
		return sendAWS(ctx, p, blacklist, blockChina, req)
	default:
		return SendResult{}, fmt.Errorf("%w: unsupported provider %s", ErrProviderUnsupported, p.Type)
	}
}

// storeRecord 单独限制进行中的请求数，pending 记录不参与完成记录的淘汰。
func (s *Service) storeRecord(r DeliveryRecord) error {
	s.recordMu.Lock()
	defer s.recordMu.Unlock()
	pending := 0
	for _, record := range s.records {
		if record.Status == "pending" {
			pending++
		}
	}
	if pending >= s.recordLimit() {
		return fmt.Errorf("%w: maximum in-flight email requests reached", ErrProviderUnavailable)
	}
	// 仅写进程内存；记录满时淘汰完成记录，不删除在途请求。
	s.records = append(s.records, r)
	s.trimRecords()
	return nil
}

// recordLimit 读取当前限额，尚无配置时使用默认值。
func (s *Service) recordLimit() int {
	if current := s.current.Load(); current != nil {
		return normalizeRecordLimit(current.RecordLimit)
	}
	return normalizeRecordLimit(0)
}

// trimRecords 只淘汰最早完成的记录，调用方须持有 recordMu。
func (s *Service) trimRecords() {
	completed := 0
	for _, record := range s.records {
		if record.Status != "pending" {
			completed++
		}
	}
	excess := completed - s.recordLimit()
	if excess <= 0 {
		return
	}
	kept := s.records[:0]
	for _, record := range s.records {
		if excess > 0 && record.Status != "pending" {
			excess--
			continue
		}
		kept = append(kept, record)
	}
	// 清空底层切片淘汰区的引用，随后缩短可见记录列表。
	clear(s.records[len(kept):])
	s.records = kept
}

// finishRecord 写入最终结果并移动到队尾，使历史记录按完成顺序保留。
func (s *Service) finishRecord(id, provider, status, providerMessageID string, accepted, attempts int, err error) {
	s.recordMu.Lock()
	defer s.recordMu.Unlock()
	for i := range s.records {
		if s.records[i].ID == id {
			s.records[i].Status = status                       // 本次发送的最终状态
			s.records[i].Provider = provider                   // 最后尝试的通道
			s.records[i].ProviderMessageID = providerMessageID // 厂商回执编号，可为空
			s.records[i].Accepted = accepted                   // 已确认受理人数
			s.records[i].Attempts = attempts                   // 所有通道累计尝试次数
			s.records[i].UpdatedAt = time.Now()                // 完成时间
			if err != nil {
				s.records[i].Error = err.Error()
			}
			// 完成后移到队尾，使淘汰顺序按完成时间排列。
			completed := s.records[i]
			copy(s.records[i:], s.records[i+1:])
			s.records[len(s.records)-1] = completed
			s.trimRecords()
			if status == "unknown" || status == "failed" {
				logx.Errorf("email submission finished delivery_id=%s provider=%s status=%s attempts=%d accepted=%d", id, provider, status, attempts, accepted)
			}
			return
		}
	}
	logx.Errorf("email submission record missing delivery_id=%s status=%s", id, status)
}

// Records 返回内存记录的副本；记录会随容量淘汰或进程退出而消失。
func (s *Service) Records() []DeliveryRecord {
	s.recordMu.Lock()
	defer s.recordMu.Unlock()
	return append([]DeliveryRecord(nil), s.records...)
}

// apply 将旧版单 SMTP 配置转换为统一配置后发布。
func (s *Service) apply(provider config.EmailProvider, blacklist []string) {
	enabled := s.enabled && hasProvider(provider)
	s.applyConfig(config.EmailConf{
		Enabled:   enabled,   // 本地开关开启且旧账号完整时才允许启用
		Provider:  provider,  // 旧版单 SMTP 配置
		Blacklist: blacklist, // 随账号更新的域名拦截列表
	})
	logx.Infof("email config applied: enabled=%t host=%s port=%s", enabled, provider.Host, smtpPort(provider))
}

// applyConfig 校验通过后一次性替换快照；远程配置不能开启本地已关闭的服务。
func (s *Service) applyConfig(c config.EmailConf) {
	c.Enabled = c.Enabled && s.enabled
	cfg, err := buildSnapshot(c)
	if err != nil {
		logx.Errorf("email config update rejected: %v", err)
		return
	}
	s.current.Store(cfg)
}

// hasProvider 仅检查 SMTP 主机和凭证是否齐全，格式由 validateProvider 校验。
func hasProvider(p config.EmailProvider) bool {
	return strings.TrimSpace(p.Host) != "" && strings.TrimSpace(p.AuthEmail) != "" && strings.TrimSpace(p.AuthPassword) != ""
}

// validateProvider 校验 SMTP 地址、认证、TLS、超时和代理参数，不建立连接。
func validateProvider(p config.EmailProvider) error {
	if !hasProvider(p) {
		return fmt.Errorf("smtp provider is incomplete")
	}
	host := strings.TrimSpace(p.Host)
	if strings.ContainsAny(host, " \t\r\n/\\") || (net.ParseIP(host) == nil && strings.Contains(host, ":")) || (net.ParseIP(host) == nil && host != "localhost" && !strings.Contains(host, ".")) {
		return fmt.Errorf("smtp host is invalid")
	}
	if !validEmail(p.AuthEmail) {
		return fmt.Errorf("smtp auth_email is invalid")
	}
	mode := strings.ToLower(strings.TrimSpace(p.TLSMode))
	if mode != "" && mode != "implicit" && mode != "starttls" {
		return fmt.Errorf("unsupported tls_mode %q", p.TLSMode)
	}
	if p.Port != "" {
		port, err := strconv.Atoi(strings.TrimSpace(p.Port))
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("smtp port is invalid")
		}
	}
	if p.TimeoutSeconds < 0 || p.TimeoutSeconds > 300 {
		return fmt.Errorf("smtp timeout_seconds must be between 0 and 300")
	}
	if p.Proxy.Type != "" || p.Proxy.Address != "" || p.Proxy.Username != "" || p.Proxy.Password != "" {
		if !strings.EqualFold(strings.TrimSpace(p.Proxy.Type), "socks5") {
			return fmt.Errorf("unsupported smtp proxy type %q", p.Proxy.Type)
		}
		proxyHost, proxyPort, err := net.SplitHostPort(p.Proxy.Address)
		if err != nil || strings.TrimSpace(proxyHost) == "" {
			return fmt.Errorf("smtp proxy address is invalid")
		}
		port, err := strconv.Atoi(proxyPort)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("smtp proxy port is invalid")
		}
	}
	return nil
}

// normalizeBlacklist 统一域名大小写并去掉开头的点，匹配时包含子域名。
func normalizeBlacklist(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(item)), ".")
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

// String 返回配置状态摘要，不包含账号密码或访问密钥。
func (s *Service) String() string {
	cfg := s.current.Load()
	if cfg == nil || !cfg.Enabled {
		return "email service not configured"
	}
	if hasProvider(cfg.Provider) {
		return fmt.Sprintf("smtp %s:%s", cfg.Provider.Host, smtpPort(cfg.Provider))
	}
	return "email service configured"
}

// newDeliveryID 为每次发送调用生成随机投递编号，不承担业务幂等校验。
func newDeliveryID() string { return "email-" + rand.Text() }
