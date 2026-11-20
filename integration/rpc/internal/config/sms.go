package config

// SMSConf 短信运行配置，包含厂商账号、通道路由、超时和 Etcd 监听参数。
type SMSConf struct {
	Enabled               bool              `json:",optional"`                          // 本地总开关，关闭后不发送短信
	Router                SMSRouterConfig   `json:",optional" yaml:"router"`            // 未指定厂商时的加权路由
	Byteplus              SMSByteplusConfig `json:",optional" yaml:"byteplus_provider"` // BytePlus 账号
	Submail               SMSSubmailConfig  `json:",optional" yaml:"submail_provider"`  // Submail 账号
	Gotone                SMSGotoneConfig   `json:",optional" yaml:"gotone_provider"`   // Gotone 账号
	AWS                   SMSAWSConfig      `json:",optional" yaml:"aws_provider"`      // AWS SNS 账号
	Aliyun                SMSAliyunConfig   `json:",optional" yaml:"ali_provider"`      // 阿里云国际短信账号
	TimeoutSeconds        int               `json:",optional"`                          // 单次厂商调用超时，秒；默认 10
	RequestTimeoutSeconds int               `json:",optional"`                          // 整批发送预算上限，秒；默认 30，另为上游截止预留响应时间
	Etcd                  SMSEtcdConfig     `json:",optional"`                          // 未填 Hosts 时仅使用本地配置
}

// SMSRouterConfig 在请求未指定 Provider 时生效，按权重生成不重复的通道顺序。
type SMSRouterConfig struct {
	Channels    []SMSRouterChannel `json:",optional" yaml:"channels"`     // 候选通道及权重
	MaxAttempts int                `json:",optional" yaml:"max_attempts"` // 最多尝试的通道数；0 表示全部候选
}

// SMSRouterChannel 表示候选通道；权重越大越可能优先选中，非正权重不参与路由。
type SMSRouterChannel struct {
	Name   string `json:",optional" yaml:"name"`   // byteplus、submail、gotone、aws、ali
	Weight int64  `json:",optional" yaml:"weight"` // 正数参与抽取，非正数跳过
}

// SMSByteplusConfig 使用 Basic Auth 调用 BytePlus 国际短信接口。
type SMSByteplusConfig struct {
	Username string `json:",optional" yaml:"username"` // Basic Auth 用户名
	Password string `json:",optional" yaml:"password"` // Basic Auth 密码
	Sender   string `json:",optional" yaml:"sender"`   // 请求中的 From 发件标识
	Endpoint string `json:",optional" yaml:"endpoint"` // 空值使用内置 BytePlus 地址
}

// SMSSubmailConfig 使用 AppID 和 AppKey 调用 Submail 国际短信接口。
type SMSSubmailConfig struct {
	AppID    string `json:",optional" yaml:"appid"`    // Submail 应用 ID
	AppKey   string `json:",optional" yaml:"app_key"`  // 请求 signature 使用的应用密钥
	Sender   string `json:",optional" yaml:"sender"`   // 可选发件标识
	Endpoint string `json:",optional" yaml:"endpoint"` // 空值使用内置国际短信地址
}

// SMSGotoneConfig 使用 Bearer Token 调用 Gotone 短信接口。
type SMSGotoneConfig struct {
	APIToken string `json:",optional" yaml:"api_token"` // Bearer 认证令牌
	SenderID string `json:",optional" yaml:"sender_id"` // 必填发件标识
	Type     string `json:",optional" yaml:"type"`      // 必填厂商短信类型
	Endpoint string `json:",optional" yaml:"endpoint"`  // 空值使用内置 Gotone 地址
}

// SMSAWSConfig 使用 AWS SNS；显式密钥与默认凭证链由配置选择。
type SMSAWSConfig struct {
	UseDefaultCredentials bool   `json:",optional" yaml:"use_default_credentials"` // 无显式密钥时，是否允许使用默认凭证链
	Region                string `json:",optional" yaml:"region"`                  // 默认 us-east-1
	AccessKeyID           string `json:",optional" yaml:"access_key_id"`           // 与 SecretAccessKey 成对填写
	SecretAccessKey       string `json:",optional" yaml:"secret_access_key"`       // 显式访问密钥
	SessionToken          string `json:",optional" yaml:"session_token"`           // 临时令牌，必须同时配置显式密钥
	SenderID              string `json:",optional" yaml:"sender_id"`               // 可选发件标识
	Type                  string `json:",optional" yaml:"type"`                    // Transactional 或 Promotional；默认前者
	MaxPrice              string `json:",optional" yaml:"max_price"`               // 可选单条短信价格上限，美元
}

// SMSAliyunConfig 使用阿里云 SendMessageToGlobe 国际短信接口，通道名为 ali。
type SMSAliyunConfig struct {
	Region          string `json:",optional" yaml:"region"`            // 默认 ap-southeast-1
	AccessKeyID     string `json:",optional" yaml:"access_key_id"`     // 阿里云访问密钥 ID
	AccessKeySecret string `json:",optional" yaml:"access_key_secret"` // 阿里云访问密钥 Secret
	Sender          string `json:",optional" yaml:"sender"`            // 请求中的 From 发件标识
	Type            string `json:",optional" yaml:"type"`              // NOTIFY 或 OTP；默认 NOTIFY
}

// SMSEtcdConfig 监听 Prefix/providers/ 下的路由和厂商配置，Prefix 默认 /sms-service。
type SMSEtcdConfig struct {
	Hosts       []string `json:",optional"` // Etcd 节点地址
	Prefix      string   `json:",optional"` // 配置前缀；缺失或删除的厂商 Key 会清空对应账号
	DialTimeout int      `json:",optional"` // Etcd 连接及读取超时，秒；默认 5
}
