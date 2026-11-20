// Package sms 提供国际手机号校验、短信通道路由和逐号码受理结果汇总。
package sms

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/nyaruka/phonenumbers"
)

var (
	ErrNotConfigured  = errors.New("sms is not configured") // 服务关闭或没有可用通道
	ErrInvalidRequest = errors.New("invalid sms request")   // 号码、正文或指定通道无效
)

// Request 是独立于 RPC 类型的短信请求。
type Request struct {
	Phones   []string // 1～100 个国际号码，发送前统一格式并去重
	Content  string   // 非空 UTF-8 正文，最多 16000 字节
	Provider string   // 空值按加权路由选择，指定时仅使用该通道
}

// RecipientResult 保存单个号码的最终结果。
// sent 表示厂商受理，failed 表示明确失败，unknown 表示无法确认受理且不应自动重发。
type RecipientResult struct {
	Phone     string `json:"phone"`                // 规范化后的 E.164 号码
	Provider  string `json:"provider"`             // 最后尝试的通道，未提交时可为空
	Status    string `json:"status"`               // sent、failed 或 unknown
	MessageID string `json:"message_id,omitempty"` // 厂商受理编号，部分通道允许为空
	Message   string `json:"message,omitempty"`    // 失败或结果不确定的原因
	Data      string `json:"data,omitempty"`       // 白名单附加信息，不透传完整响应
}

// SendResult 汇总本次调用结果，不持久化，也不代表手机最终收到短信。
type SendResult struct {
	Accepted   int               // 厂商受理的唯一号码数
	ErrorCode  int32             // 0 全部受理、2 失败或部分完成、3 存在不确定结果
	Message    string            // 整批结果说明
	DeliveryID string            // 本批随机编号，不是幂等键
	Results    []RecipientResult // 每个去重号码的最终结果
	Errors     []string          // 最终失败或不确定号码的原因
	Data       string            // Results 的 JSON 字符串，兼容旧调用方
}

// outcome 是厂商适配层的单号码结果，供上层决定是否切换通道。
type outcome struct {
	status    string // sent 已受理、failed 明确失败、unknown 结果不确定
	messageID string // 厂商受理编号
	message   string // 经筛选的结果说明
	data      string // 经筛选的附加数据
}

// Service 使用原子配置快照编排发送，并复用 HTTP 连接及 AWS 客户端。
type Service struct {
	current atomic.Pointer[config.SMSConf] // 当前只读配置，每批发送固定使用一份
	enabled bool                           // 本地总开关，远程配置不能覆盖
	closed  atomic.Bool                    // 服务是否已关闭
	watcher *watcher                       // 可选的 Etcd 配置监听器
	client  *http.Client                   // 共用连接池的 HTTP 客户端

	send func(context.Context, config.SMSConf, string, string, string) outcome // 单号码发送函数

	ctx       context.Context     // 服务生命周期，控制发送取消
	cancel    context.CancelFunc  // 服务关闭时取消在途请求
	closeOnce sync.Once           // 保证资源只释放一次
	awsMu     sync.Mutex          // 保护 AWS 客户端缓存
	awsConfig config.SMSAWSConfig // 构建缓存客户端时使用的配置
	awsClient snsAPI              // 复用的 SNS 客户端，配置变更后重建
}

// New 校验配置并启动可选的 Etcd 监听，初始化过程不向厂商发送短信。
func New(c config.SMSConf) (*Service, error) {
	cfg, err := normalizeConfig(c)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = base.Clone()
	}
	ctx, cancel := context.WithCancel(context.Background())
	service := &Service{
		enabled: c.Enabled, // 保存不可被远程覆盖的本地开关
		ctx:     ctx,       // 统一控制请求取消
		cancel:  cancel,    // 关闭服务时触发
		// 禁止自动跟随重定向，避免将带凭证的发送请求转发到其他地址。
		client: &http.Client{
			Transport:     transport,                                                                     // 复用连接池与环境代理设置
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, // 返回首个响应
		},
	}
	service.current.Store(&cfg)
	service.send = service.sendProvider
	if c.Enabled && len(c.Etcd.Hosts) > 0 {
		// 仅在本地启用并配置 Etcd 时启动后台监听，发送请求读取原子快照。
		service.watcher, err = newWatcher(c.Etcd, cfg, service.apply)
		if err != nil {
			cancel()
			service.client.CloseIdleConnections()
			return nil, err
		}
		go service.watcher.run()
	}
	return service, nil
}

// Close 取消发送、停止配置监听并关闭空闲连接，可重复调用。
func (s *Service) Close() error {
	if s != nil {
		s.closeOnce.Do(func() {
			s.closed.Store(true)
			s.cancel()
			if s.watcher != nil {
				s.watcher.close()
			}
			s.client.CloseIdleConnections()
		})
	}
	return nil
}

// apply 校验成功后原子替换配置，远程配置不能绕过本地 Enabled 总开关。
func (s *Service) apply(c config.SMSConf) error {
	c.Enabled = c.Enabled && s.enabled
	cfg, err := normalizeConfig(c)
	if err != nil {
		return err
	}
	s.current.Store(&cfg)
	return nil
}

// Send 先校验整批号码，再逐通道发送；已开始处理的批次通过结果字段表达部分失败。
func (s *Service) Send(ctx context.Context, req Request) (SendResult, error) {
	if s == nil || s.closed.Load() {
		return SendResult{}, ErrNotConfigured
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}
	// 服务关闭时也取消在途请求；本批发送固定使用同一份配置。
	requestCtx, cancel := context.WithCancel(ctx)
	stopCancel := context.AfterFunc(s.ctx, cancel)
	defer stopCancel()
	defer cancel()
	ctx = requestCtx
	cfg := s.current.Load()
	if cfg == nil || !cfg.Enabled {
		return SendResult{}, ErrNotConfigured
	}
	// 整批请求共用预算，并在 RPC 或调用方截止前预留结果汇总和传输时间。
	// 每次厂商调用仍受 TimeoutSeconds 限制，超时后的 unknown 号码不再重发。
	budgetCtx, stopBudget := context.WithTimeout(ctx, requestBudget(ctx, time.Duration(cfg.RequestTimeoutSeconds)*time.Second))
	defer stopBudget()
	ctx = budgetCtx
	// 全部号码校验通过并选好候选通道后，才开始任何厂商调用。
	phones, err := normalizePhones(req)
	if err != nil {
		return SendResult{}, err
	}
	channels, err := channelOrder(*cfg, req.Provider)
	if err != nil {
		return SendResult{}, err
	}
	var identifier [16]byte
	if _, err := rand.Read(identifier[:]); err != nil {
		return SendResult{}, err
	}
	result := SendResult{
		DeliveryID: hex.EncodeToString(identifier[:]),    // 本批随机编号，不写数据库
		Results:    make([]RecipientResult, len(phones)), // 每个去重号码单独保留最终结果
	}
	for index, phone := range phones {
		result.Results[index] = RecipientResult{
			Phone:   phone,                      // 规范化后的国际号码
			Status:  "failed",                   // 尚未提交时不计为受理成功
			Message: "no available sms channel", // 提交后由厂商结果覆盖
		}
	}
	for _, channel := range channels {
		for index := range result.Results {
			item := &result.Results[index]
			// 只对明确失败的号码尝试下一通道，受理或状态不明的号码均不再提交。
			if item.Status == "sent" || item.Status == "unknown" {
				continue
			}
			// 预算耗尽时保留已有结果，未继续提交的号码保持失败状态。
			if ctx.Err() != nil || s.closed.Load() {
				for remaining := range result.Results {
					if result.Results[remaining].Status == "failed" {
						result.Results[remaining].Message = "sms request stopped before further submissions"
					}
				}
				return finishResult(result), nil
			}
			sendCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutSeconds)*time.Second)
			sent := s.send(sendCtx, *cfg, channel, item.Phone, req.Content)
			cancel()
			item.Provider = channel         // 记录本次实际尝试的厂商
			item.Status = sent.status       // 决定是否允许尝试下一通道
			item.MessageID = sent.messageID // 厂商受理编号
			item.Message = sent.message     // 筛选后的失败或不确定原因
			item.Data = sent.data           // 厂商白名单附加数据
		}
	}
	return finishResult(result), nil
}

// requestBudget 在上游剩余时间内预留 10%（最多 1 秒），避免部分发送结果被外层超时丢弃。
// 未设置上游截止时间时只使用配置预算；预留时间不能保证断网或主动取消时仍能收到响应。
func requestBudget(ctx context.Context, configured time.Duration) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		reserve := remaining / 10
		if reserve > time.Second {
			reserve = time.Second
		}
		if available := remaining - reserve; available < configured {
			return available
		}
	}
	return configured
}

// finishResult 只根据最终状态重建统计和 Errors，去掉已切换成功号码的历史错误。
func finishResult(result SendResult) SendResult {
	// 重新统计最终结果，不保留已通过备用通道成功号码的中间失败记录。
	result.Accepted = 0
	result.Errors = nil
	uncertain := false
	for _, item := range result.Results {
		if item.Status == "sent" {
			result.Accepted++
		} else {
			// failed 和 unknown 都保留原因，另记是否存在无法确认受理的号码。
			result.Errors = append(result.Errors, fmt.Sprintf("%s (%s): %s", item.Phone, item.Provider, item.Message))
			uncertain = uncertain || item.Status == "unknown"
		}
	}
	result.Message = "send completed"
	// 任一号码未明确受理时返回 2；全部受理时沿用结果的默认成功码 0。
	if result.Accepted != len(result.Results) {
		result.ErrorCode = 2
		result.Message = "sms send failed or partially completed"
	}
	if uncertain {
		// 不确定结果优先于普通失败，提示调用方禁止自动重发 unknown 号码。
		result.ErrorCode = 3
		result.Message = "sms delivery is uncertain; do not resend unknown recipients automatically"
	}
	// 同时返回逐号码列表和其 JSON 字符串，兼容读取旧 Data 字段的调用方。
	data, _ := json.Marshal(result.Results)
	result.Data = string(data)
	return result
}

// normalizePhones 在任何发送前完成整批校验，统一为 E.164 格式后去重。
// 号码结构有效不代表该号码已开通或可接收短信。
func normalizePhones(req Request) ([]string, error) {
	if len(req.Phones) == 0 || len(req.Phones) > 100 || strings.TrimSpace(req.Content) == "" || !utf8.ValidString(req.Content) || len(req.Content) > 16000 {
		return nil, fmt.Errorf("%w: 1-100 phones and nonempty UTF-8 content up to 16000 bytes are required", ErrInvalidRequest)
	}
	// 清理空格和连字符，再按国际号码规则校验；任一号码无效则整批拒绝。
	phones := make([]string, 0, len(req.Phones))
	seen := make(map[string]bool)
	for _, phone := range req.Phones {
		phone = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(phone))
		if !strings.HasPrefix(phone, "+") {
			phone = "+" + phone
		}
		if len(phone) < 3 || len(phone) > 16 || strings.Trim(phone[1:], "0123456789") != "" {
			return nil, fmt.Errorf("%w: invalid international phone", ErrInvalidRequest)
		}
		number, err := phonenumbers.Parse(phone, "")
		if err != nil || !phonenumbers.IsValidNumber(number) {
			return nil, fmt.Errorf("%w: invalid international phone", ErrInvalidRequest)
		}
		normalized := phonenumbers.Format(number, phonenumbers.E164)
		if !seen[normalized] {
			phones = append(phones, normalized)
			seen[normalized] = true
		}
	}
	return phones, nil
}

// channelOrder 支持显式指定单通道，或按正权重抽取不重复的候选通道顺序。
// MaxAttempts 限制候选通道数，不是在同一通道内重试的次数。
func channelOrder(c config.SMSConf, selected string) ([]string, error) {
	selected = strings.ToLower(strings.TrimSpace(selected))
	if selected == "aliyun" {
		selected = "ali"
	}
	// 显式指定时只用该通道，不因失败自动切换到其他厂商。
	if selected != "" {
		if !supportedChannel(selected) {
			return nil, fmt.Errorf("%w: unsupported sms provider", ErrInvalidRequest)
		}
		if err := validateChannel(c, selected); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrNotConfigured, err)
		}
		return []string{selected}, nil
	}
	// 自动路由只纳入正权重且账号完整的通道，没有候选时不调用任何厂商。
	candidates := make([]config.SMSRouterChannel, 0, len(c.Router.Channels))
	for _, channel := range c.Router.Channels {
		if channel.Weight > 0 && validateChannel(c, channel.Name) == nil {
			candidates = append(candidates, channel)
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: sms router has no configured positive channel", ErrNotConfigured)
	}
	maxAttempts := c.Router.MaxAttempts
	if maxAttempts <= 0 || maxAttempts > len(candidates) {
		maxAttempts = len(candidates)
	}
	// 每轮按剩余权重抽取一个通道并移除，形成不重复的候选顺序。
	order := make([]string, 0, maxAttempts)
	for len(order) < maxAttempts {
		total := int64(0)
		for _, channel := range candidates {
			total += channel.Weight
		}
		pick, err := rand.Int(rand.Reader, big.NewInt(total))
		if err != nil {
			return nil, err
		}
		remaining := pick.Int64()
		for index, channel := range candidates {
			if remaining < channel.Weight {
				order = append(order, channel.Name)
				candidates = append(candidates[:index], candidates[index+1:]...)
				break
			}
			remaining -= channel.Weight
		}
	}
	return order, nil
}
