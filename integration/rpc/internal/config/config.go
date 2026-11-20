package config

import "github.com/zeromicro/go-zero/zrpc"

// Config RPC 服务配置。
type Config struct {
	zrpc.RpcServerConf
	DB      DBConf      `json:",optional"` // 业务库（Provider / 缓存表 / 本地 IP）
	IpGeo   IpGeoConf   `json:",optional"` // IP/GEO 查询与种子配置
	Message MessageConf `json:",optional"` // 短信、邮件等消息能力
}

// MessageConf 邮件、短信配置；配置了 Etcd 时在后台加载并监听更新。
type MessageConf struct {
	Email EmailConf `json:",optional"` // 邮件通道与发送策略
	SMS   SMSConf   `json:",optional"` // 短信通道与路由策略
}

// EmailConf 邮件运行配置。
type EmailConf struct {
	Enabled         bool                  `json:",optional"`                          // 本地总开关，关闭后远程配置不能开启
	Provider        EmailProvider         `json:",optional"`                          // 单 SMTP 配置，Providers 为空时使用
	Providers       []EmailProviderConfig `json:",optional"`                          // 多通道配置
	Blacklist       []string              `json:",optional"`                          // 拒绝这些域名及其子域名的收件人
	BlockChinaEmail bool                  `json:",optional" yaml:"block_china_email"` // 拦截内置国内邮箱域名列表
	Etcd            EmailEtcdConfig       `json:",optional"`                          // 远程通道配置，未填 Hosts 时仅用本地配置
	Retry           EmailRetryConfig      `json:",optional"`                          // 明确可重试错误的退避策略
	RecordLimit     int                   `json:",optional"`                          // 完成记录数、在途请求数分别限额；默认 100，最大 10000
}

// EmailProvider SMTP 连接配置，可通过本地 YAML 或 Etcd 的 smtp_provider 设置。
type EmailProvider struct {
	Host           string           `json:",optional" yaml:"host"`            // SMTP 主机，不带协议或端口
	Port           string           `json:",optional" yaml:"port"`            // 空值按 TLS 模式使用 465 或 587
	AuthEmail      string           `json:",optional" yaml:"auth_email"`      // 登录邮箱，同时作为发件地址
	AuthPassword   string           `json:",optional" yaml:"auth_password"`   // SMTP 密码或授权码
	TLSMode        string           `json:",optional" yaml:"tls_mode"`        // implicit 或 starttls；空值为 implicit
	TLSSkipVerify  bool             `json:",optional" yaml:"tls_skip_verify"` // 跳过服务器证书校验，TLS 加密仍启用
	TimeoutSeconds int              `json:",optional" yaml:"timeout_seconds"` // 秒，0 使用连接 5 秒、发送 10 秒的默认值
	Proxy          EmailProxyConfig `json:",optional" yaml:"proxy"`           // 可选 SOCKS5 代理
}

// EmailProviderConfig 描述一个发信通道，Type 支持 smtp、aliyun、aws（ses 别名）。
type EmailProviderConfig struct {
	Name     string            `json:",optional" yaml:"name"`     // 唯一通道名，空值使用 Type
	Type     string            `json:",optional" yaml:"type"`     // smtp、aliyun、aws；ses 视为 aws
	Enabled  bool              `json:",optional" yaml:"enabled"`  // 当前通道是否参与发送
	Priority int               `json:",optional" yaml:"priority"` // 越小越优先，相同值保持配置顺序
	SMTP     EmailProvider     `json:",optional" yaml:"smtp"`     // Type=smtp 时使用
	Aliyun   AliyunEmailConfig `json:",optional" yaml:"aliyun"`   // Type=aliyun 时使用
	AWS      AWSEmailConfig    `json:",optional" yaml:"aws"`      // Type=aws 时使用
}

// AliyunEmailConfig 阿里云 DirectMail 配置，AccountName 为已配置的发信地址。
type AliyunEmailConfig struct {
	Region          string `json:",optional" yaml:"region"`            // DirectMail 区域
	AccessKeyID     string `json:",optional" yaml:"access_key_id"`     // 阿里云访问密钥 ID
	AccessKeySecret string `json:",optional" yaml:"access_key_secret"` // 阿里云访问密钥 Secret
	AccountName     string `json:",optional" yaml:"account_name"`      // 已配置的发信地址
	FromAlias       string `json:",optional" yaml:"from_alias"`        // 可选发件人显示名
}

// AWSEmailConfig AWS SES 配置；不填写显式密钥时使用 SDK 默认凭证链。
type AWSEmailConfig struct {
	Region          string `json:",optional" yaml:"region"`            // SES 区域
	AccessKeyID     string `json:",optional" yaml:"access_key_id"`     // 与 SecretAccessKey 成对填写
	SecretAccessKey string `json:",optional" yaml:"secret_access_key"` // 显式访问密钥
	SessionToken    string `json:",optional" yaml:"session_token"`     // 显式临时凭证的会话令牌
	From            string `json:",optional" yaml:"from"`              // SES 发件地址
}

// EmailProxyConfig 为 SMTP 连接提供可选的 SOCKS5 代理。
type EmailProxyConfig struct {
	Type     string `json:",optional" yaml:"type"`     // socks5
	Address  string `json:",optional" yaml:"address"`  // 代理地址，格式为 host:port
	Username string `json:",optional" yaml:"username"` // 可选代理用户名
	Password string `json:",optional" yaml:"password"` // 可选代理密码
}

// EmailRetryConfig 控制明确可重试错误的指数退避；结果不确定时不执行重试。
type EmailRetryConfig struct {
	MaxAttempts  int `json:",optional" yaml:"max_attempts"`   // 每通道尝试次数，包含首次；默认 3，最大 10
	BackoffMS    int `json:",optional" yaml:"backoff_ms"`     // 初始退避毫秒数；默认 100
	MaxBackoffMS int `json:",optional" yaml:"max_backoff_ms"` // 退避上限毫秒数；默认 5000，最大 60000
}

// EmailEtcdConfig 邮件远程配置；Key 缺失或删除时停用邮件通道。
type EmailEtcdConfig struct {
	Hosts       []string `json:",optional"` // Etcd 节点地址
	Key         string   `json:",optional"` // 默认 /integration/message/email/providers/smtp
	DialTimeout int      `json:",optional"` // 连接超时秒数；默认 5，配置读取固定 10 秒
}

// DBConf PostgreSQL 连接。
type DBConf struct {
	Driver string `json:",optional"` // 仅支持 postgres；空则按 postgres
	DSN    string `json:",optional"` // 必填；库不存在时启动会自动创建
}

// IpGeoConf IP/GEO 运行时配置。
type IpGeoConf struct {
	CacheTTL    int64          `json:",optional"` // 缓存小时数；写入 expires_at；0=不过期
	CacheDriver string         `json:",optional"` // table | memory | redis；空则 table
	CacheRedis  RedisConf      `json:",optional"` // CacheDriver=redis 时必填 Host
	Providers   []ProviderConf `json:",optional"` // ip_geo_provider 为空时的全局种子
}

// RedisConf Redis 缓存连接（CacheDriver=redis）。
type RedisConf struct {
	Host      string `json:",optional"` // 地址，如 127.0.0.1:6379
	Pass      string `json:",optional"` // 密码，可空
	DB        int    `json:",optional"` // 库号
	KeyPrefix string `json:",optional"` // 键前缀；空则 ipgeo
}

// ProviderConf 全局定位源种子（写入 ip_geo_provider；已有数据不覆盖）。
type ProviderConf struct {
	Name       string            `json:",optional"` // 源名称；缓存键前缀
	Driver     string            `json:",optional"` // http | mmdb | ip2region | local
	Tag        string            `json:",optional"` // 展示标签：在线 | 本地 | 离线
	IsEnabled  bool              `json:",optional"` // 默认是否启用（可被租户策略覆盖）
	Priority   int               `json:",optional"` // 默认优先级，越小越优先
	Endpoint   string            `json:",optional"` // HTTP URL 模板或本地库路径；local 可空
	AuthType   string            `json:",optional"` // none | basic | bearer | header
	AccessKey  string            `json:",optional"` // 访问密钥 / basic 用户名
	SecretKey  string            `json:",optional"` // 密钥 / basic 密码
	TimeoutSec int               `json:",optional"` // HTTP 超时秒数
	FieldMap   map[string]string `json:",optional"` // 我方字段 -> 三方 JSON 路径（gjson）
	Headers    map[string]string `json:",optional"` // 额外 HTTP 请求头
}
