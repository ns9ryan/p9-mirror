package email

import (
	"fmt"
	"regexp"
	"strings"

	"oa.98ent.com/p9/integration/rpc/internal/config"
)

// emailRegionPattern 限制厂商区域名称的格式，不检查该区域是否实际开通。
var emailRegionPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`)

// buildSnapshot 统一规范化本地和 Etcd 配置，复制通道列表后构造只读快照。
// 禁用的通道保留名称占位，但不要求填写完整的厂商参数。
func buildSnapshot(c config.EmailConf) (*snapshot, error) {
	if !c.Enabled {
		return &snapshot{
			Enabled:     false,                               // 禁用时不要求填写厂商凭证
			Retry:       normalizeRetry(c.Retry),             // 保留规范化重试策略
			RecordLimit: normalizeRecordLimit(c.RecordLimit), // 保留内存限额
		}, nil
	}
	c.Provider = normalizeSMTPProvider(c.Provider)
	if smtpConfigured(c.Provider) {
		if err := validateProvider(c.Provider); err != nil {
			return nil, err
		}
	}
	// 复制通道配置后统一名称和类型，避免修改已发布快照中的列表。
	providers := append([]config.EmailProviderConfig(nil), c.Providers...)
	names := make(map[string]bool, len(providers)) // 检查通道名称是否重复
	for index := range providers {
		provider := &providers[index]
		// 清理连接参数和发件地址，不改变访问密钥的原始内容。
		provider.SMTP = normalizeSMTPProvider(provider.SMTP)                         // 统一 SMTP 地址、端口和 TLS 模式
		provider.Aliyun.Region = strings.TrimSpace(provider.Aliyun.Region)           // 阿里云区域
		provider.Aliyun.AccountName = strings.TrimSpace(provider.Aliyun.AccountName) // 阿里云发信地址
		provider.AWS.Region = strings.TrimSpace(provider.AWS.Region)                 // AWS SES 区域
		provider.AWS.From = strings.TrimSpace(provider.AWS.From)                     // AWS SES 发信地址
		provider.Type = strings.ToLower(strings.TrimSpace(provider.Type))            // 通道类型统一小写
		if provider.Type == "" {
			provider.Type = "smtp" // 未指定类型时兼容 SMTP
		}
		if provider.Type == "ses" {
			provider.Type = "aws" // SES 别名统一为 AWS
		}
		provider.Name = strings.TrimSpace(provider.Name) // 保留名称大小写，支持精确选择
		if provider.Name == "" {
			provider.Name = provider.Type // 未命名时使用通道类型
		}
		// 包括禁用通道在内，名称都必须唯一，避免指定名称时产生歧义。
		if names[provider.Name] {
			return nil, fmt.Errorf("duplicate email provider name %q", provider.Name)
		}
		names[provider.Name] = true
		// 禁用通道只保留名称与配置，不要求账号完整，也不参与后续发送。
		if !provider.Enabled {
			continue
		}
		var err error
		// 启用通道按厂商校验配置；校验失败则拒绝整份候选快照。
		switch provider.Type {
		case "smtp":
			err = validateProvider(provider.SMTP)
		case "aliyun":
			err = validateAliyunConfig(provider.Aliyun)
		case "aws":
			err = validateAWSConfig(provider.AWS)
		default:
			err = fmt.Errorf("unsupported email provider type %q", provider.Type)
		}
		if err != nil {
			return nil, fmt.Errorf("email provider %s: %w", provider.Name, err)
		}
	}
	return &snapshot{
		Provider:        c.Provider,                          // 规范化后的旧版 SMTP 配置
		Providers:       providers,                           // 校验后的独立通道列表
		Enabled:         c.Enabled,                           // 本次配置的启用状态
		Blacklist:       normalizeBlacklist(c.Blacklist),     // 域名统一大小写并去除空项
		BlockChinaEmail: c.BlockChinaEmail,                   // 内置域名拦截开关
		Retry:           normalizeRetry(c.Retry),             // 补齐默认值并限制重试上限
		RecordLimit:     normalizeRecordLimit(c.RecordLimit), // 限制记录与在途请求数量
	}, nil
}

// normalizeSMTPProvider 清理地址和模式字段，不修改密码等凭证内容。
func normalizeSMTPProvider(provider config.EmailProvider) config.EmailProvider {
	provider.Host = strings.TrimSpace(provider.Host)                              // SMTP 主机，不含协议和端口
	provider.Port = strings.TrimSpace(provider.Port)                              // 显式端口，允许留空使用默认值
	provider.AuthEmail = strings.TrimSpace(provider.AuthEmail)                    // 登录邮箱兼作发件地址
	provider.TLSMode = strings.ToLower(strings.TrimSpace(provider.TLSMode))       // 统一 TLS 模式
	provider.Proxy.Type = strings.ToLower(strings.TrimSpace(provider.Proxy.Type)) // 统一代理类型
	provider.Proxy.Address = strings.TrimSpace(provider.Proxy.Address)            // SOCKS5 代理地址
	return provider
}

// validateAliyunConfig 检查区域、密钥和发件地址，不调用厂商接口。
func validateAliyunConfig(c config.AliyunEmailConfig) error {
	if !emailRegionPattern.MatchString(c.Region) || strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.AccessKeySecret) == "" || !validEmail(c.AccountName) {
		return fmt.Errorf("aliyun credentials, region and a valid account_name are required")
	}
	return nil
}

// validateAWSConfig 检查 SES 区域和发件地址，显式密钥必须成对填写。
func validateAWSConfig(c config.AWSEmailConfig) error {
	if !emailRegionPattern.MatchString(c.Region) || !validEmail(c.From) {
		return fmt.Errorf("aws region and a valid from are required")
	}
	if (c.AccessKeyID == "") != (c.SecretAccessKey == "") {
		return fmt.Errorf("both aws access_key_id and secret_access_key are required")
	}
	if c.AccessKeyID != "" && (strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.SecretAccessKey) == "") {
		return fmt.Errorf("aws credentials cannot contain only whitespace")
	}
	return nil
}
