package geo

import "time"

// LangNames 多语言名称，key 为语言码（en / zh-CN 等）。
type LangNames map[string]string

// Record 查询中间结果，序列化进缓存 payload。
type Record struct {
	IP              string    `json:"ip"`                        // 规范化后的 IP
	ContinentCode   string    `json:"continent_code"`            // 大洲代码，如 AS
	ContinentNames  LangNames `json:"continent_names,omitempty"` // 大洲多语言名
	CountryCode     string    `json:"country_code"`              // 国家/地区 ISO 码
	CountryNames    LangNames `json:"country_names,omitempty"`   // 国家多语言名
	ProvinceIsoCode string    `json:"province_iso_code"`         // 省/州 ISO 码
	ProvinceNames   LangNames `json:"province_names,omitempty"`  // 省/州多语言名
	CityNames       LangNames `json:"city_names,omitempty"`      // 城市多语言名
	PostalCode      string    `json:"postal_code"`               // 邮编
	Latitude        float64   `json:"latitude"`                  // 纬度
	Longitude       float64   `json:"longitude"`                 // 经度
	Timezone        string    `json:"timezone"`                  // 时区，如 Asia/Shanghai
	Isp             string    `json:"isp"`                       // ISP
	Org             string    `json:"org"`                       // 组织
	ASN             string    `json:"asn"`                       // ASN
	Source          string    `json:"source"`                    // 命中源名称（Provider.Name）
	OperatorCode    string    `json:"operator_code,omitempty"`   // 查询时租户编码
	ExpiresAt       time.Time `json:"expires_at,omitempty"`      // 缓存过期时间
	UpdatedAt       time.Time `json:"updated_at,omitempty"`      // 最近写入/刷新时间
}

// Info 面向 RPC/HTTP 的单条结果（已按 lang 选好地名）。
type Info struct {
	Query           string  // 原始查询 IP/域名
	Continent       string  // 大洲名（按语言）
	ContinentCode   string  // 大洲代码
	Country         string  // 国家名（按语言）
	CountryCode     string  // 国家/地区 ISO 码
	Province        string  // 省/州名（按语言）
	ProvinceIsoCode string  // 省/州 ISO 码
	City            string  // 城市名（按语言）
	Zip             string  // 邮编（与 Postal 同源）
	Lat             float64 // 纬度
	Lon             float64 // 经度
	Timezone        string  // 时区
	Isp             string  // ISP
	Org             string  // 组织
	As              string  // ASN
	Postal          string  // 邮编
	Source          string  // 命中源名称
}

// ToInfo 按语言选取地名，转为对外 Info。
func (r *Record) ToInfo(lang, query string) *Info {
	if r == nil {
		return &Info{Query: query}
	}
	return &Info{
		Query:           query,
		Continent:       pickName(r.ContinentNames, lang),
		ContinentCode:   r.ContinentCode,
		Country:         pickName(r.CountryNames, lang),
		CountryCode:     r.CountryCode,
		Province:        pickName(r.ProvinceNames, lang),
		ProvinceIsoCode: r.ProvinceIsoCode,
		City:            pickName(r.CityNames, lang),
		Zip:             r.PostalCode,
		Lat:             r.Latitude,
		Lon:             r.Longitude,
		Timezone:        r.Timezone,
		Isp:             r.Isp,
		Org:             r.Org,
		As:              r.ASN,
		Postal:          r.PostalCode,
		Source:          r.Source,
	}
}
