package geo

import (
	"strings"
	"time"
)

// LocalIP 本地库一条 IP 地理信息。
type LocalIP struct {
	ID              int64   // 主键
	IP              string  // 规范化 IP（唯一）
	Continent       string  // 大洲名
	ContinentCode   string  // 大洲代码
	Country         string  // 国家名
	CountryCode     string  // 国家/地区 ISO 码
	Province        string  // 省/州名
	ProvinceIsoCode string  // 省/州 ISO 码
	City            string  // 城市名
	Postal          string  // 邮编
	Lat             float64 // 纬度
	Lon             float64 // 经度
	Timezone        string  // 时区
	Isp             string  // ISP
	Org             string  // 组织
	ASN             string  // ASN
}

// LocalIPPage 本地 IP 分页结果。
type LocalIPPage struct {
	List  []LocalIP // 当前页数据，不为 nil
	Total int       // 符合筛选的总数
}

// NormalizePage 规范化分页（默认 page=1、page_size=20，最大 100）。
func NormalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// NormalizeLocalIP 校验并规范化 LocalIP 字段。
func NormalizeLocalIP(in LocalIP) (LocalIP, error) {
	raw := strings.TrimSpace(in.IP)
	if raw == "" {
		return in, ErrEmptyIP
	}
	ip, err := NormalizeIP(raw)
	if err != nil {
		return in, ErrBadIP
	}
	in.IP = ip
	in.Continent = strings.TrimSpace(in.Continent)
	in.ContinentCode = strings.TrimSpace(in.ContinentCode)
	in.Country = strings.TrimSpace(in.Country)
	in.CountryCode = strings.TrimSpace(in.CountryCode)
	in.Province = strings.TrimSpace(in.Province)
	in.ProvinceIsoCode = strings.TrimSpace(in.ProvinceIsoCode)
	in.City = strings.TrimSpace(in.City)
	in.Postal = strings.TrimSpace(in.Postal)
	in.Timezone = strings.TrimSpace(in.Timezone)
	in.Isp = strings.TrimSpace(in.Isp)
	in.Org = strings.TrimSpace(in.Org)
	in.ASN = strings.TrimSpace(in.ASN)
	return in, nil
}

// ToRecord 转为查询 Record。
func (in LocalIP) ToRecord(source string) *Record {
	return &Record{
		IP:              in.IP,
		ContinentCode:   in.ContinentCode,
		ContinentNames:  localNames(in.Continent),
		CountryCode:     in.CountryCode,
		CountryNames:    localNames(in.Country),
		ProvinceIsoCode: in.ProvinceIsoCode,
		ProvinceNames:   localNames(in.Province),
		CityNames:       localNames(in.City),
		PostalCode:      in.Postal,
		Latitude:        in.Lat,
		Longitude:       in.Lon,
		Timezone:        in.Timezone,
		Isp:             in.Isp,
		Org:             in.Org,
		ASN:             in.ASN,
		Source:          source,
		UpdatedAt:       time.Now(),
	}
}

func localNames(v string) LangNames {
	if v == "" {
		return nil
	}
	return LangNames{"en": v, "zh-CN": v}
}
