package sms

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"oa.98ent.com/p9/integration/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

// watcher 维护 providers 前缀下的路由和厂商配置，单个 Key 更新失败时保留其原值。
type watcher struct {
	client    *clientv3.Client           // 配置读取与监听连接
	prefix    string                     // 实际监听的 providers 路径前缀
	config    config.SMSConf             // 最近一次成功发布的配置基准
	apply     func(config.SMSConf) error // 校验并发布完整候选配置
	ctx       context.Context            // 控制读取、监听和重试退出
	cancel    context.CancelFunc         // 关闭时取消后台工作
	done      chan struct{}              // run 退出时关闭，供 close 等待
	timeout   time.Duration              // Etcd 连接及每次读取超时
	closeOnce sync.Once                  // 保证关闭过程只执行一次
}

// newWatcher 创建前缀监听器，远程读取和更新由 run 在后台执行。
func newWatcher(c config.SMSEtcdConfig, base config.SMSConf, apply func(config.SMSConf) error) (*watcher, error) {
	// 清理空白节点，过滤后至少保留一个 Etcd 地址。
	hosts := make([]string, 0, len(c.Hosts))
	for _, host := range c.Hosts {
		if host = strings.TrimSpace(host); host != "" {
			hosts = append(hosts, host)
		}
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("sms etcd hosts are required")
	}
	// 空前缀使用默认目录，统一去掉末尾斜杠；根路径和相对路径均拒绝。
	prefix := strings.TrimSpace(c.Prefix)
	if prefix == "" {
		prefix = "/sms-service"
	}
	prefix = strings.TrimRight(prefix, "/")
	if !strings.HasPrefix(prefix, "/") {
		return nil, fmt.Errorf("sms etcd prefix must be a non-root absolute path")
	}
	// 未配置或非正数时使用 5 秒，超过 120 秒则拒绝创建监听器。
	seconds := c.DialTimeout
	if seconds <= 0 {
		seconds = 5
	}
	if seconds > 120 {
		return nil, fmt.Errorf("sms etcd dial timeout exceeds 120 seconds")
	}
	timeout := time.Duration(seconds) * time.Second
	// 此处创建客户端，实际读取和后续断线恢复由后台 run 处理。
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,   // 去除空项后的 Etcd 节点
		DialTimeout: timeout, // 建立连接的超时
	})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{
		client:  client,                 // 共用读取与监听连接
		prefix:  prefix + "/providers/", // 配置根路径下的厂商与路由目录
		config:  base,                   // 首次远程读取前使用的本地配置
		apply:   apply,                  // 有效配置发布回调
		ctx:     ctx,                    // 后台工作生命周期
		cancel:  cancel,                 // 服务关闭时取消
		done:    make(chan struct{}),    // 监听协程退出通知
		timeout: timeout,                // 每次配置读取的超时
	}, nil
}

// run 先读取全部 Key，再从下一版本监听；缺失 Key 会清空对应本地账号或路由。
// 断线后每隔 1 秒重读，读取失败保留原配置，直到服务关闭。
func (w *watcher) run() {
	defer close(w.done)
	for w.ctx.Err() == nil {
		// 首次启动和监听重置后都先读取完整前缀，每次读取受 timeout 限制。
		ctx, cancel := context.WithTimeout(w.ctx, w.timeout)
		response, err := w.client.Get(ctx, w.prefix, clientv3.WithPrefix())
		cancel()
		if err == nil {
			// 按相对 Key 建立快照；未出现的受管 Key 按删除处理。
			values := make(map[string][]byte)
			for _, value := range response.Kvs {
				values[strings.TrimPrefix(string(value.Key), w.prefix)] = value.Value
			}
			for _, name := range []string{"router", "byteplus", "submail", "gotone", "aws", "ali"} {
				w.update(name, values[name], values[name] == nil)
			}
			// 从快照的下一修订版本开始监听，衔接读取期间发生的配置变化。
			watchCtx, stopWatch := context.WithCancel(w.ctx)
			watch := w.client.Watch(watchCtx, w.prefix, clientv3.WithPrefix(), clientv3.WithRev(response.Header.Revision+1))
			for event := range watch {
				if event.Err() != nil {
					// 监听失败后退出本轮，等待后重新读取完整快照。
					break
				}
				// 更新只替换对应 Key；删除会清空对应账号或路由。
				for _, change := range event.Events {
					if change.Kv != nil {
						name := strings.TrimPrefix(string(change.Kv.Key), w.prefix)
						w.update(name, change.Kv.Value, change.Type == clientv3.EventTypeDelete)
					}
				}
			}
			stopWatch()
		} else if w.ctx.Err() == nil {
			logx.Error("sms etcd configuration read failed; retaining last valid configuration")
		}
		// 读失败或监听结束后等待 1 秒再重建，关闭服务时立即停止等待。
		timer := time.NewTimer(time.Second)
		select {
		case <-w.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

// update 在解析、校验及发布都成功后才更新本地基准，不记录包含账号的原始配置。
func (w *watcher) update(name string, data []byte, deleted bool) {
	// 忽略目录内不属于短信路由或厂商配置的 Key。
	if name != "router" && !supportedChannel(name) {
		return
	}
	// 候选配置通过解析和发布后才更新基准；失败不覆盖最近一次有效值。
	next, err := parseProvider(w.config, name, data, deleted)
	if err == nil {
		err = w.apply(next)
	}
	if err != nil {
		logx.Errorf("sms etcd invalid configuration for %s; retaining last valid value", name)
		return
	}
	w.config = next
}

// close 取消读取和监听、关闭连接，并等待已启动的 run 退出。
func (w *watcher) close() {
	w.closeOnce.Do(func() {
		w.cancel()
		_ = w.client.Close()
		<-w.done
	})
}

// parseProvider 按 Key 对应的 YAML 根字段解析单个文档，拒绝未知字段。
// 删除只清空对应通道或路由，再对整份候选配置执行规范化和校验。
func parseProvider(base config.SMSConf, name string, data []byte, deleted bool) (config.SMSConf, error) {
	var value struct {
		Router   *config.SMSRouterConfig   `yaml:"router"`            // 权重和候选通道数量限制
		Byteplus *config.SMSByteplusConfig `yaml:"byteplus_provider"` // Byteplus 配置，nil 表示未提交该根字段
		Submail  *config.SMSSubmailConfig  `yaml:"submail_provider"`  // Submail 账号与发送地址
		Gotone   *config.SMSGotoneConfig   `yaml:"gotone_provider"`   // Gotone 令牌与发送参数
		AWS      *config.SMSAWSConfig      `yaml:"aws_provider"`      // SNS 凭证、区域和发送参数
		Aliyun   *config.SMSAliyunConfig   `yaml:"ali_provider"`      // 阿里云凭证、区域和发送参数
	}
	if !deleted {
		// 普通更新要求严格的单文档 YAML；删除事件无需解析旧内容。
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(&value); err != nil {
			return base, fmt.Errorf("invalid sms YAML")
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return base, fmt.Errorf("sms YAML must contain one document")
		}
	}
	// 只替换当前 Key 对应的配置；删除时清空，未涉及的通道保留原值。
	switch name {
	case "router":
		base.Router = config.SMSRouterConfig{}
		if !deleted {
			if value.Router == nil {
				return base, fmt.Errorf("router root is required")
			}
			base.Router = *value.Router
		}
	case "byteplus":
		base.Byteplus = config.SMSByteplusConfig{}
		if !deleted {
			if value.Byteplus == nil {
				return base, fmt.Errorf("byteplus_provider root is required")
			}
			base.Byteplus = *value.Byteplus
		}
	case "submail":
		base.Submail = config.SMSSubmailConfig{}
		if !deleted {
			if value.Submail == nil {
				return base, fmt.Errorf("submail_provider root is required")
			}
			base.Submail = *value.Submail
		}
	case "gotone":
		base.Gotone = config.SMSGotoneConfig{}
		if !deleted {
			if value.Gotone == nil {
				return base, fmt.Errorf("gotone_provider root is required")
			}
			base.Gotone = *value.Gotone
		}
	case "aws":
		base.AWS = config.SMSAWSConfig{}
		if !deleted {
			if value.AWS == nil {
				return base, fmt.Errorf("aws_provider root is required")
			}
			base.AWS = *value.AWS
		}
	case "ali":
		base.Aliyun = config.SMSAliyunConfig{}
		if !deleted {
			if value.Aliyun == nil {
				return base, fmt.Errorf("ali_provider root is required")
			}
			base.Aliyun = *value.Aliyun
		}
	default:
		return base, fmt.Errorf("unsupported sms configuration key")
	}
	// 替换后检查整份配置，避免一个 Key 破坏其他通道的约束。
	normalized, err := normalizeConfig(base)
	if err != nil {
		return base, err
	}
	if !deleted && name != "router" {
		// 新提交的厂商配置必须完整，空账号文档不能当作有效更新发布。
		if err := validateChannel(normalized, name); err != nil {
			return base, err
		}
	}
	return normalized, nil
}
