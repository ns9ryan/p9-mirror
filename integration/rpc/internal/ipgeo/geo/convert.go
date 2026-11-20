package geo

import (
	"errors"
	"net"
)

var supportedLang = map[string]bool{
	"en":    true,
	"zh-CN": true,
	"pt-BR": true,
	"ru":    true,
	"fr":    true,
	"ja":    true,
	"es":    true,
	"de":    true,
}

// SupportedLang 是否支持的查询语言。
func SupportedLang(lang string) bool {
	return supportedLang[lang]
}

func pickName(names LangNames, lang string) string {
	if names == nil {
		return ""
	}
	if v, ok := names[lang]; ok && v != "" {
		return v
	}
	if v, ok := names["en"]; ok {
		return v
	}
	for _, v := range names {
		return v
	}
	return ""
}

// NormalizeIP 规范化 IPv4/IPv6 字符串。
func NormalizeIP(ip string) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", errors.New("invalid ip")
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String(), nil
	}
	return parsed.String(), nil
}

// IsPrivate 是否私网/回环/链路本地。
func IsPrivate(ip net.IP) bool {
	return ip != nil && (ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast())
}

// Truncate 截断字符串到最多 n 字节。
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// PrivateInfo 私网 IP 的固定 Info。
func PrivateInfo(query string) *Info {
	return &Info{
		Query:   query,
		Country: "私有IP",
		Source:  SourcePrivate,
	}
}
