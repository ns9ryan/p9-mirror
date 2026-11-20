// Package ipgeo 对 logic/svc 暴露查询引擎入口（re-export lookup + 常用 geo 类型/错误）。
//
// 子包职责：
//   - lookup：Search 编排
//   - source：http / mmdb / ip2region / local 驱动
//   - cache：查询结果缓存（table / memory / redis）
//   - geo：领域类型、校验、租户编码
//
// 管理端 CRUD 不在本包，见 rpc/service。
package ipgeo

import (
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/geo"
	"oa.98ent.com/p9/integration/rpc/internal/ipgeo/lookup"
)

// 查询入口与领域类型别名。
type (
	Lookup      = lookup.Lookup
	Conf        = lookup.Conf
	Provider    = geo.Provider
	Info        = geo.Info
	LocalIP     = geo.LocalIP
	PolicyItem  = geo.PolicyItem
	LocalIPPage = geo.LocalIPPage
)

// New 构造 Lookup；AccountOperatorCode 从登录态 / metadata 解析租户。
var (
	New                 = lookup.New
	AccountOperatorCode = geo.AccountOperatorCode
)

// 领域错误（logic 映射 gRPC status 用）。
var (
	ErrBadDriver         = geo.ErrBadDriver
	ErrBadProvider       = geo.ErrBadProvider
	ErrEmptyName         = geo.ErrEmptyName
	ErrEmptyEndpoint     = geo.ErrEmptyEndpoint
	ErrNameTaken         = geo.ErrNameTaken
	ErrEmptyIP           = geo.ErrEmptyIP
	ErrBadIP             = geo.ErrBadIP
	ErrIPTaken           = geo.ErrIPTaken
	ErrEmptyPolicy       = geo.ErrEmptyPolicy
	ErrTooManyPolicies   = geo.ErrTooManyPolicies
	ErrDuplicateProvider = geo.ErrDuplicateProvider
	ErrBadTag            = geo.ErrBadTag
	ErrNotFound          = geo.ErrNotFound
	ErrLocalIPNotFound   = geo.ErrLocalIPNotFound
	ErrTooManyIPs        = geo.ErrTooManyIPs
	ErrBadLang           = geo.ErrBadLang
	ErrMissingDSN        = geo.ErrMissingDSN
)

// DefaultLang 查询默认语言。
const DefaultLang = geo.DefaultLang
