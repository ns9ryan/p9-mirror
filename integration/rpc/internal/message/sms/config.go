package sms

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"oa.98ent.com/p9/integration/rpc/internal/config"
)

// normalizeConfig 统一默认值、路由权重和厂商参数，供启动与热更新共用。
// 只要求已配置的厂商参数完整，允许未使用的厂商暂不填写账号。
func normalizeConfig(c config.SMSConf) (config.SMSConf, error) {
	// 单次厂商调用默认 10 秒，整批默认 30 秒；显式超出范围则拒绝配置。
	if c.TimeoutSeconds == 0 {
		c.TimeoutSeconds = 10
	}
	if c.TimeoutSeconds < 1 || c.TimeoutSeconds > 120 {
		return config.SMSConf{}, fmt.Errorf("sms timeout must be between 1 and 120 seconds")
	}
	if c.RequestTimeoutSeconds == 0 {
		c.RequestTimeoutSeconds = 30
	}
	if c.RequestTimeoutSeconds < 1 || c.RequestTimeoutSeconds > 300 {
		return config.SMSConf{}, fmt.Errorf("sms request timeout must be between 1 and 300 seconds")
	}
	if len(c.Router.Channels) > 100 {
		return config.SMSConf{}, fmt.Errorf("invalid sms router limits")
	}
	if c.Router.MaxAttempts < 0 {
		// 非正数表示不额外限制候选通道数，实际数量由可用通道决定。
		c.Router.MaxAttempts = 0
	}
	// 创建独立路由切片，合并重复通道并限制权重总和，不修改已发布的快照。
	positions := make(map[string]int)
	channels := make([]config.SMSRouterChannel, 0, len(c.Router.Channels))
	totalWeight := int64(0)
	for _, channel := range c.Router.Channels {
		channel.Name = strings.ToLower(strings.TrimSpace(channel.Name))
		if channel.Weight <= 0 {
			continue
		}
		if !supportedChannel(channel.Name) || channel.Weight > 1000000000000 {
			return config.SMSConf{}, fmt.Errorf("unsupported sms router channel or excessive weight")
		}
		totalWeight += channel.Weight
		if totalWeight > 1000000000000 {
			return config.SMSConf{}, fmt.Errorf("sms router total weight exceeds limit")
		}
		if index, exists := positions[channel.Name]; exists {
			channels[index].Weight += channel.Weight
		} else {
			positions[channel.Name] = len(channels)
			channels = append(channels, channel)
		}
	}
	c.Router.Channels = channels
	c.Byteplus.Username = strings.TrimSpace(c.Byteplus.Username) // Byteplus 认证账号
	c.Submail.AppID = strings.TrimSpace(c.Submail.AppID)         // Submail 应用编号
	c.Submail.AppKey = strings.TrimSpace(c.Submail.AppKey)       // Submail 应用密钥
	c.Gotone.APIToken = strings.TrimSpace(c.Gotone.APIToken)     // Gotone 认证令牌
	c.Gotone.SenderID = strings.TrimSpace(c.Gotone.SenderID)     // Gotone 发件标识
	c.Gotone.Type = strings.TrimSpace(c.Gotone.Type)             // Gotone 短信类型
	// 空值补默认配置，仅更新内存候选值，不写回本地文件或 Etcd。
	c.Byteplus.Endpoint = defaultEndpoint(c.Byteplus.Endpoint, "https://sms.byteplusapi.com/sms/openapi/send_sms")      // Byteplus 发送地址
	c.Submail.Endpoint = defaultEndpoint(c.Submail.Endpoint, "https://api-v4.mysubmail.com/internationalsms/send.json") // Submail 发送地址
	c.Gotone.Endpoint = defaultEndpoint(c.Gotone.Endpoint, "https://app.gotonesms.net/api/v3/sms/send")                 // Gotone 发送地址

	c.AWS.Region = defaultEndpoint(c.AWS.Region, "us-east-1")            // SNS 区域
	c.AWS.Type = defaultEndpoint(c.AWS.Type, "Transactional")            // 默认事务短信
	c.Aliyun.Region = defaultEndpoint(c.Aliyun.Region, "ap-southeast-1") // 国际短信区域
	c.Aliyun.Type = defaultEndpoint(c.Aliyun.Type, "NOTIFY")             // 默认通知短信
	// 三个 HTTP 通道的地址始终校验，关闭服务也不接受非法发送地址。
	for _, endpoint := range []string{c.Byteplus.Endpoint, c.Submail.Endpoint, c.Gotone.Endpoint} {
		if err := validateEndpoint(endpoint); err != nil {
			return config.SMSConf{}, err
		}
	}
	if !c.Enabled {
		// 总开关关闭时允许账号未配置完整，发送入口仍会拒绝调用。
		return c, nil
	}
	// 启用时只检查已填写账号的通道，未使用的厂商可留空。
	for _, name := range []string{"byteplus", "submail", "gotone", "aws", "ali"} {
		if configured(c, name) {
			if err := validateChannel(c, name); err != nil {
				return config.SMSConf{}, err
			}
		}
	}
	return c, nil
}

// defaultEndpoint 去掉首尾空白，空值使用默认值；也用于区域和短信类型字段。
func defaultEndpoint(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

// validateEndpoint 要求 HTTPS；仅允许本机环回地址使用 HTTP，供本地模拟测试。
func validateEndpoint(endpoint string) error {
	// 拒绝缺少主机、URL 内嵌账号或片段，避免凭证和请求地址混用。
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("invalid sms endpoint")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	// HTTP 例外只识别 localhost 和环回 IP，不接受普通域名或内网地址。
	address := net.ParseIP(parsed.Hostname())
	if parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || address != nil && address.IsLoopback()) {
		return nil
	}
	return fmt.Errorf("sms endpoint must use HTTPS; HTTP is only allowed for local tests")
}

// supportedChannel 判断内部通道名称；请求中的 aliyun 别名由 channelOrder 转换。
func supportedChannel(name string) bool {
	switch name {
	case "byteplus", "submail", "gotone", "aws", "ali":
		return true
	default:
		return false
	}
}

// configured 判断是否开始配置通道；完整性检查交给 validateChannel。
func configured(c config.SMSConf, name string) bool {
	// 任一认证字段已填写即视为开始配置，不能因凭证不完整而跳过校验。
	switch name {
	case "byteplus":
		return c.Byteplus.Username != "" || c.Byteplus.Password != ""
	case "submail":
		return c.Submail.AppID != "" || c.Submail.AppKey != ""
	case "gotone":
		return c.Gotone.APIToken != ""
	case "aws":
		return c.AWS.UseDefaultCredentials || c.AWS.AccessKeyID != "" || c.AWS.SecretAccessKey != "" || c.AWS.SenderID != "" || c.AWS.SessionToken != ""
	case "ali":
		return c.Aliyun.AccessKeyID != "" || c.Aliyun.AccessKeySecret != ""
	}
	return false
}

// validateChannel 检查发送所需参数；AWS 默认凭证链必须通过配置显式启用。
func validateChannel(c config.SMSConf, name string) error {
	switch name {
	case "byteplus":
		if c.Byteplus.Username == "" || c.Byteplus.Password == "" {
			return fmt.Errorf("byteplus username and password are required")
		}
	case "submail":
		if c.Submail.AppID == "" || c.Submail.AppKey == "" {
			return fmt.Errorf("submail appid and app_key are required")
		}
	case "gotone":
		if c.Gotone.APIToken == "" || strings.TrimSpace(c.Gotone.SenderID) == "" || strings.TrimSpace(c.Gotone.Type) == "" {
			return fmt.Errorf("gotone api_token, sender_id and type are required")
		}
	case "aws":
		// 默认凭证链需显式允许；填写静态密钥时，ID 和密钥必须成对出现。
		if c.AWS.AccessKeyID == "" && !c.AWS.UseDefaultCredentials {
			return fmt.Errorf("aws explicit credentials or use_default_credentials are required")
		}
		if (c.AWS.AccessKeyID == "") != (c.AWS.SecretAccessKey == "") {
			return fmt.Errorf("aws access_key_id and secret_access_key must be set together")
		}
		// 临时令牌只能配合显式密钥使用，不能单独作为认证凭证。
		if c.AWS.SessionToken != "" && c.AWS.AccessKeyID == "" {
			return fmt.Errorf("aws session_token requires explicit credentials")
		}
		// 保留厂商要求的大小写，区分事务短信和营销短信。
		if c.AWS.Type != "Transactional" && c.AWS.Type != "Promotional" {
			return fmt.Errorf("aws sms type must be Transactional or Promotional")
		}
	case "ali":
		if c.Aliyun.AccessKeyID == "" || c.Aliyun.AccessKeySecret == "" {
			return fmt.Errorf("ali access_key_id and access_key_secret are required")
		}
		// 国际短信仅接受通知、验证码或营销三种类型。
		if c.Aliyun.Type != "NOTIFY" && c.Aliyun.Type != "OTP" && c.Aliyun.Type != "PROMOTION" {
			return fmt.Errorf("invalid ali sms type")
		}
	default:
		return fmt.Errorf("unsupported sms provider")
	}
	return nil
}
