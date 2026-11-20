package geo

import (
	"context"
	"errors"
)

// 驱动、鉴权、标签与语言常量。
const (
	DriverHTTP      = "http"
	DriverMMDB      = "mmdb"
	DriverIP2Region = "ip2region"
	DriverLocal     = "local"

	AuthNone   = "none"
	AuthBasic  = "basic"
	AuthBearer = "bearer"
	AuthHeader = "header"

	SourcePrivate = "private"

	TagOnline  = "在线"
	TagLocal   = "本地"
	TagOffline = "离线"

	MaxPolicies = 100
	DefaultLang = "en"
)

// 领域错误。
var (
	ErrTooManyIPs        = errors.New("max ips is 100")
	ErrBadLang           = errors.New("unsupported lang")
	ErrNotFound          = errors.New("provider not found")
	ErrBadDriver         = errors.New("unsupported driver")
	ErrBadProvider       = errors.New("invalid provider")
	ErrEmptyName         = errors.New("provider name is required")
	ErrEmptyEndpoint     = errors.New("provider endpoint is required")
	ErrNameTaken         = errors.New("provider name already exists")
	ErrMissingDSN        = errors.New("DB.DSN is required")
	ErrLocalIPNotFound   = errors.New("local ip not found")
	ErrEmptyIP           = errors.New("ip is required")
	ErrBadIP             = errors.New("invalid ip")
	ErrIPTaken           = errors.New("ip already exists")
	ErrEmptyPolicy       = errors.New("policy list is required")
	ErrTooManyPolicies   = errors.New("max policies is 100")
	ErrDuplicateProvider = errors.New("duplicate provider_id")
	ErrBadTag            = errors.New("unsupported tag")
	ErrBadCacheDriver    = errors.New("unsupported cache driver")
	ErrMissingRedis      = errors.New("IpGeo.CacheRedis.Host is required")
)

// Provider 全局定位源。IsEnabled/Priority 在按租户读取时可能被策略覆盖。
type Provider struct {
	ID           int64             // 主键
	Name         string            // 源名称；缓存键前缀
	Driver       string            // 驱动：http / mmdb / ip2region / local
	Tag          string            // 展示标签（如 maxmind）
	IsEnabled    bool              // 是否启用（可被租户策略覆盖）
	Priority     int               // 越小越优先（可被租户策略覆盖）
	Endpoint     string            // HTTP URL 模板或本地库文件路径
	AuthType     string            // 鉴权：none / basic / bearer / header
	AccessKey    string            // 访问密钥 / 用户名
	SecretKey    string            // 密钥 / 密码
	TimeoutSec   int               // HTTP 超时秒数
	OperatorCode string            // 查询时租户编码（策略叠加后写入，非 yaml 种子字段）
	FieldMap     map[string]string // 响应 JSON 路径 → Record 字段
	Headers      map[string]string // 额外 HTTP 请求头
}

// Source 运行时查询源。
type Source interface {
	Fetch(ctx context.Context, ip, lang string) (*Record, error)
	Close() error
}

// PolicyItem 租户策略项（启用与优先级）。
type PolicyItem struct {
	ProviderID int64 // 全局 Provider 主键
	IsEnabled  bool  // 该租户是否启用此源
	Priority   int   // 该租户下的优先级，越小越优先
}
