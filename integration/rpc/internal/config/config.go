package config

import "github.com/zeromicro/go-zero/zrpc"

// Config RPC 服务配置。
type Config struct {
	zrpc.RpcServerConf
	DB    DBConf    `json:",optional"` // 业务库（Provider / 缓存表 / 本地 IP）
	IpGeo IpGeoConf `json:",optional"` // IP/GEO 查询与种子配置
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
	Name         string            `json:",optional"` // 源名称；缓存键前缀
	Driver       string            `json:",optional"` // http | mmdb | ip2region | local
	Tag          string            `json:",optional"` // 展示标签：在线 | 本地 | 离线
	IsEnabled    bool              `json:",optional"` // 默认是否启用（可被租户策略覆盖）
	Priority     int               `json:",optional"` // 默认优先级，越小越优先
	Endpoint     string            `json:",optional"` // HTTP URL 模板或本地库路径；local 可空
	AuthType   string            `json:",optional"` // none | basic | bearer | header
	AccessKey  string            `json:",optional"` // 访问密钥 / basic 用户名
	SecretKey  string            `json:",optional"` // 密钥 / basic 密码
	TimeoutSec int               `json:",optional"` // HTTP 超时秒数
	FieldMap   map[string]string `json:",optional"` // 我方字段 -> 三方 JSON 路径（gjson）
	Headers    map[string]string `json:",optional"` // 额外 HTTP 请求头
}
